# AWS Hosted Control Plane: Provider/Client Setup

This doc guide you how to install hosted cluster and use provider/client model on aws.

> Prerequisite: A host cluster requires at least 6 nodes
> (3 control plane and 3 worker nodes) with `m5.4xlarge` instance type.

### Install multicluster engine

Create multicluster engine operator

```bash
cat <<EOF | oc apply -f -
---
apiVersion: v1
kind: Namespace
metadata:
  name: multicluster-engine
---
apiVersion: operators.coreos.com/v1
kind: OperatorGroup
metadata:
  name: multicluster-engine-abcde
  namespace: multicluster-engine
spec:
  targetNamespaces:
  - multicluster-engine
  upgradeStrategy: Default
---
apiVersion: operators.coreos.com/v1alpha1
kind: Subscription
metadata:
  name: multicluster-engine
  namespace: multicluster-engine
spec:
  installPlanApproval: Automatic
  name: multicluster-engine
  source: redhat-operators
  sourceNamespace: openshift-marketplace
EOF
```

Create the MultiClusterEngine CR

```bash
cat <<EOF | oc apply -f -
apiVersion: multicluster.openshift.io/v1
kind: MultiClusterEngine
metadata:
  name: multiclusterengine
spec: {}
EOF
```

Verify MCE installation

```bash
oc get mce
oc get managedclusters
oc get pods -n multicluster-engine
```

---

### Create an S3 Bucket for OIDC Discovery

```bash
REGION=us-east-1
BUCKET=nigoyal-hcp-oidc-bucket
```

```bash
if [ "$REGION" = "us-east-1" ]; then
  aws s3api create-bucket \
    --region "$REGION" \
    --bucket "$BUCKET"
else
  aws s3api create-bucket \
    --region "$REGION" \
    --bucket "$BUCKET" \
    --create-bucket-configuration LocationConstraint="$REGION"
fi

aws s3api put-public-access-block \
  --region "$REGION" \
  --bucket "$BUCKET" \
  --public-access-block-configuration '{
    "BlockPublicAcls": true,
    "IgnorePublicAcls": true,
    "BlockPublicPolicy": false,
    "RestrictPublicBuckets": false
  }'

aws s3api put-bucket-policy \
  --region "$REGION" \
  --bucket "$BUCKET" \
  --policy '{
    "Version": "2012-10-17",
    "Statement": [
      {
        "Effect": "Allow",
        "Principal": "*",
        "Action": "s3:GetObject",
        "Resource": "arn:aws:s3:::'"$BUCKET"'/*"
      }
    ]
  }'

aws s3api get-bucket-policy \
  --region "$REGION" \
  --bucket "$BUCKET" \
  --query Policy \
  --output text | jq .
```

---

### Create the OIDC Secret on the Management Cluster

Create a credentials file with your AWS access key and secret.

```bash
cat <<EOF > aws-creds
[default]
aws_access_key_id = <id>
aws_secret_access_key = <key>
EOF
```

Create the secret (used by the HyperShift operator).

```bash
REGION=us-east-1
BUCKET=nigoyal-hcp-oidc-bucket
```

```bash
oc create secret generic hypershift-operator-oidc-provider-s3-credentials \
  --from-file=credentials=./aws-creds \
  --from-literal=region="$REGION" \
  --from-literal=bucket="$BUCKET" \
  -n local-cluster
```

---

### Create an IAM Role for Cluster Provisioning

```bash
ROLE=nigoyal-hcp-cli-role
```

```bash
ARN=$(aws sts get-caller-identity --query "Arn" --output text)

aws iam create-role \
  --role-name "$ROLE" \
  --query "Role.Arn" \
  --assume-role-policy-document '{
    "Version": "2012-10-17",
    "Statement": [
      {
        "Effect": "Allow",
        "Principal": {
          "AWS": "'"$ARN"'"
        },
        "Action": "sts:AssumeRole"
      }
    ]
  }'
```

Attach the provisioning policy. This policy grants the permissions to create and tear down hosted-cluster infrastructure on AWS.

```bash
aws iam put-role-policy \
  --role-name "$ROLE" \
  --policy-name "$ROLE"-policy \
  --policy-document '{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "EC2",
      "Effect": "Allow",
      "Action": [
        "ec2:AllocateAddress",
        "ec2:AssociateDhcpOptions",
        "ec2:AssociateRouteTable",
        "ec2:AttachInternetGateway",
        "ec2:CreateDhcpOptions",
        "ec2:CreateInternetGateway",
        "ec2:CreateNatGateway",
        "ec2:CreateRoute",
        "ec2:CreateRouteTable",
        "ec2:CreateSubnet",
        "ec2:CreateTags",
        "ec2:CreateVpc",
        "ec2:CreateVpcEndpoint",
        "ec2:DeleteDhcpOptions",
        "ec2:DeleteInternetGateway",
        "ec2:DeleteNatGateway",
        "ec2:DeleteRoute",
        "ec2:DeleteRouteTable",
        "ec2:DeleteSecurityGroup",
        "ec2:DeleteSubnet",
        "ec2:DeleteVpc",
        "ec2:DeleteVpcEndpoints",
        "ec2:DeleteVpcEndpointServiceConfigurations",
        "ec2:DescribeAddresses",
        "ec2:DescribeAvailabilityZones",
        "ec2:DescribeDhcpOptions",
        "ec2:DescribeInstances",
        "ec2:DescribeInternetGateways",
        "ec2:DescribeNatGateways",
        "ec2:DescribeRouteTables",
        "ec2:DescribeSecurityGroups",
        "ec2:DescribeSubnets",
        "ec2:DescribeVpcEndpointConnections",
        "ec2:DescribeVpcEndpoints",
        "ec2:DescribeVpcEndpointServiceConfigurations",
        "ec2:DescribeVpcs",
        "ec2:DetachInternetGateway",
        "ec2:DisassociateRouteTable",
        "ec2:ModifyVpcAttribute",
        "ec2:RejectVpcEndpointConnections",
        "ec2:ReleaseAddress",
        "ec2:ReplaceRouteTableAssociation",
        "ec2:RevokeSecurityGroupEgress",
        "ec2:RevokeSecurityGroupIngress",
        "ec2:TerminateInstances"
      ],
      "Resource": "*"
    },
    {
      "Sid": "ELB",
      "Effect": "Allow",
      "Action": [
        "elasticloadbalancing:DeleteLoadBalancer",
        "elasticloadbalancing:DeleteTargetGroup",
        "elasticloadbalancing:DescribeLoadBalancers",
        "elasticloadbalancing:DescribeTargetGroups"
      ],
      "Resource": "*"
    },
    {
      "Sid": "IAM",
      "Effect": "Allow",
      "Action": [
        "iam:AddRoleToInstanceProfile",
        "iam:CreateInstanceProfile",
        "iam:CreateOpenIDConnectProvider",
        "iam:CreateRole",
        "iam:DeleteInstanceProfile",
        "iam:DeleteOpenIDConnectProvider",
        "iam:DeleteRole",
        "iam:DeleteRolePolicy",
        "iam:DetachRolePolicy",
        "iam:GetInstanceProfile",
        "iam:GetRole",
        "iam:GetRolePolicy",
        "iam:ListAttachedRolePolicies",
        "iam:ListInstanceProfilesForRole",
        "iam:ListOpenIDConnectProviders",
        "iam:ListRolePolicies",
        "iam:PutRolePolicy",
        "iam:RemoveRoleFromInstanceProfile",
        "iam:TagRole",
        "iam:UpdateAssumeRolePolicy",
        "iam:UpdateRole"
      ],
      "Resource": "*"
    },
    {
      "Sid": "Route53",
      "Effect": "Allow",
      "Action": [
        "route53:AssociateVPCWithHostedZone",
        "route53:ChangeResourceRecordSets",
        "route53:CreateHostedZone",
        "route53:DeleteHostedZone",
        "route53:ListHostedZones",
        "route53:ListHostedZonesByName",
        "route53:ListHostedZonesByVPC",
        "route53:ListResourceRecordSets"
      ],
      "Resource": "*"
    },
    {
      "Sid": "S3",
      "Effect": "Allow",
      "Action": [
        "s3:DeleteBucket",
        "s3:DeleteObject",
        "s3:ListAllMyBuckets",
        "s3:ListBucket"
      ],
      "Resource": "*"
    },
    {
      "Sid": "IAMPassRole",
      "Effect": "Allow",
      "Action": "iam:PassRole",
      "Resource": "arn:*:iam::*:role/*-worker-role",
      "Condition": {
        "ForAnyValue:StringEqualsIfExists": {
          "iam:PassedToService": "ec2.amazonaws.com"
        }
      }
    }
  ]
}'
```

---

### Create the Hosted Cluster

```bash
REGION=us-east-1
ROLE=nigoyal-hcp-cli-role
BASE_DOMAIN=ocs.syseng.devcluster.openshift.com
HOSTED_VPC_CIDR=10.1.0.0/16
```

```bash
aws sts get-session-token --output json > sts-creds.json

ARN=$(aws iam get-role --role-name "$ROLE" --query 'Role.Arn' --output text)

hcp create cluster aws \
  --name hosted \
  --infra-id hosted \
  --namespace hosted \
  --region "$REGION" \
  --base-domain "$BASE_DOMAIN" \
  --sts-creds ./sts-creds.json \
  --pull-secret ./secrets.json \
  --generate-ssh \
  --node-pool-replicas 3 \
  --vpc-cidr "$HOSTED_VPC_CIDR" \
  --role-arn "$ARN"
```

Verify the hosted cluster

```bash
oc get hostedclusters -n hosted
oc get nodepools -n hosted
oc get pods -n hosted-hosted
```

Generate a kubeconfig and confirm the hosted cluster is operational.

```bash
hcp create kubeconfig --name hosted --namespace hosted > kubeconfig.hosted

oc --kubeconfig kubeconfig.hosted get clusterversion
oc --kubeconfig kubeconfig.hosted get nodes
oc --kubeconfig kubeconfig.hosted get co
```

---

### Create vpc peering

Create a vpc peering connection b/w host and hosted cluster to enable communication.

```bash
REGION=us-east-1
KUBECONFIG_HOST=kubeconfig.host
KUBECONFIG_HOSTED=kubeconfig.hosted
```

```bash
HOST_INSTANCE_ID=$(oc get nodes -l node-role.kubernetes.io/worker \
  -o jsonpath='{.items[0].spec.providerID}' \
  --kubeconfig "$KUBECONFIG_HOST" | xargs basename)

HOSTED_INSTANCE_ID=$(oc get nodes -l node-role.kubernetes.io/worker \
  -o jsonpath='{.items[0].spec.providerID}' \
  --kubeconfig "$KUBECONFIG_HOSTED" | xargs basename)

HOST_VPC_ID=$(aws ec2 describe-instances \
  --region "$REGION" \
  --instance-ids "$HOST_INSTANCE_ID" \
  --query 'Reservations[0].Instances[0].VpcId' \
  --output text)

HOSTED_VPC_ID=$(aws ec2 describe-instances \
  --region "$REGION" \
  --instance-ids "$HOSTED_INSTANCE_ID" \
  --query 'Reservations[0].Instances[0].VpcId' \
  --output text)

PCX_ID=$(aws ec2 create-vpc-peering-connection \
  --region "$REGION" \
  --vpc-id "$HOSTED_VPC_ID" \
  --peer-vpc-id "$HOST_VPC_ID" \
  --query 'VpcPeeringConnection.VpcPeeringConnectionId' \
  --output text)

aws ec2 accept-vpc-peering-connection \
  --region "$REGION" \
  --vpc-peering-connection-id "$PCX_ID" \
  --output off
```

---

### Create route for host cluster

Create route for host cluster to enable communication from host to hosted cluster.

```bash
REGION=us-east-1
KUBECONFIG_HOST=kubeconfig.host
HOSTED_VPC_CIDR=10.1.0.0/16
```

```bash
for HOST_INSTANCE_ID in $(oc get nodes -l node-role.kubernetes.io/worker \
  -o json --kubeconfig "$KUBECONFIG_HOST" | jq -r '
  .items |
  group_by(.metadata.labels["topology.kubernetes.io/zone"])[] |
  .[0].spec.providerID |
  split("/")[-1]
'); do

  HOST_SUBNET_ID=$(aws ec2 describe-instances \
    --region "$REGION" \
    --instance-ids "$HOST_INSTANCE_ID" \
    --query 'Reservations[0].Instances[0].SubnetId' \
    --output text)

  HOST_ROUTE_TABLE_ID=$(aws ec2 describe-route-tables \
    --region "$REGION" \
    --filters "Name=association.subnet-id,Values=$HOST_SUBNET_ID" \
    --query 'RouteTables[0].RouteTableId' \
    --output text)

  HOST_VPC_ID=$(aws ec2 describe-instances \
    --region "$REGION" \
    --instance-ids "$HOST_INSTANCE_ID" \
    --query 'Reservations[0].Instances[0].VpcId' \
    --output text)

  PCX_ID=$(aws ec2 describe-vpc-peering-connections \
    --region "$REGION" \
    --filters "Name=accepter-vpc-info.vpc-id,Values=$HOST_VPC_ID" \
    --query 'VpcPeeringConnections[].VpcPeeringConnectionId' \
    --output text)

  aws ec2 create-route \
    --region "$REGION" \
    --route-table-id "$HOST_ROUTE_TABLE_ID" \
    --destination-cidr-block "$HOSTED_VPC_CIDR" \
    --vpc-peering-connection-id "$PCX_ID"
done
```

---

### Create route for hosted cluster

Create route for hosted cluster to enable communication from hosted to host cluster.

```bash
REGION=us-east-1
KUBECONFIG_HOSTED=kubeconfig.hosted
HOST_VPC_CIDR=10.0.0.0/16
```

```bash
for HOSTED_INSTANCE_ID in $(oc get nodes -l node-role.kubernetes.io/worker \
  -o json --kubeconfig "$KUBECONFIG_HOSTED" | jq -r '
  .items |
  group_by(.metadata.labels["topology.kubernetes.io/zone"])[] |
  .[0].spec.providerID |
  split("/")[-1]
'); do

  HOSTED_SUBNET_ID=$(aws ec2 describe-instances \
    --region "$REGION" \
    --instance-ids "$HOSTED_INSTANCE_ID" \
    --query 'Reservations[0].Instances[0].SubnetId' \
    --output text)

  HOSTED_ROUTE_TABLE_ID=$(aws ec2 describe-route-tables \
    --region "$REGION" \
    --filters "Name=association.subnet-id,Values=$HOSTED_SUBNET_ID" \
    --query 'RouteTables[0].RouteTableId' \
    --output text)

  HOSTED_VPC_ID=$(aws ec2 describe-instances \
    --region "$REGION" \
    --instance-ids "$HOSTED_INSTANCE_ID" \
    --query 'Reservations[0].Instances[0].VpcId' \
    --output text)

  PCX_ID=$(aws ec2 describe-vpc-peering-connections \
    --region "$REGION" \
    --filters "Name=requester-vpc-info.vpc-id,Values=$HOSTED_VPC_ID" \
    --query 'VpcPeeringConnections[].VpcPeeringConnectionId' \
    --output text)

  aws ec2 create-route \
    --region "$REGION" \
    --route-table-id "$HOSTED_ROUTE_TABLE_ID" \
    --destination-cidr-block "$HOST_VPC_CIDR" \
    --vpc-peering-connection-id "$PCX_ID"
done
```

---

### Enable odf ports

Add odf ports to the aws host cluster security group.

```bash
REGION=us-east-1
KUBECONFIG_HOST=kubeconfig.host
HOST_VPC_CIDR=10.0.0.0/16
HOSTED_VPC_CIDR=10.1.0.0/16
```

```bash
HOST_INSTANCE_ID=$(oc get nodes -l node-role.kubernetes.io/worker \
  -o jsonpath='{.items[0].spec.providerID}' \
  --kubeconfig "$KUBECONFIG_HOST" | xargs basename)

HOST_SG_ID=$(aws ec2 describe-instances \
  --region "$REGION" \
  --instance-ids "$HOST_INSTANCE_ID" \
  --query 'Reservations[0].Instances[0].SecurityGroups[0].GroupId' \
  --output text)

aws ec2 authorize-security-group-ingress \
  --region "$REGION" \
  --group-id "$HOST_SG_ID" \
  --ip-permissions '[
    {
      "IpProtocol": "tcp",
      "FromPort": 31659,
      "ToPort": 31659,
      "IpRanges": [{"CidrIp": "'"$HOST_VPC_CIDR"'"}]
    },
    {
     "IpProtocol": "tcp",
      "FromPort": 3300,
      "ToPort": 3300,
      "IpRanges": [{"CidrIp": "'"$HOST_VPC_CIDR"'"}]
    },
    {
      "IpProtocol": "tcp",
      "FromPort": 6789,
      "ToPort": 6789,
      "IpRanges": [{"CidrIp": "'"$HOST_VPC_CIDR"'"}]
    },
    {
      "IpProtocol": "tcp",
      "FromPort": 9283,
      "ToPort": 9283,
      "IpRanges": [{"CidrIp": "'"$HOST_VPC_CIDR"'"}]
    },
    {
      "IpProtocol": "tcp",
      "FromPort": 6800,
      "ToPort": 7300,
      "IpRanges": [{"CidrIp": "'"$HOST_VPC_CIDR"'"}]
    }
  ]' --output off

aws ec2 authorize-security-group-ingress \
  --region "$REGION" \
  --group-id "$HOST_SG_ID" \
  --ip-permissions '[
    {
      "IpProtocol": "tcp",
      "FromPort": 31659,
      "ToPort": 31659,
      "IpRanges": [{"CidrIp": "'"$HOSTED_VPC_CIDR"'"}]
    },
    {
     "IpProtocol": "tcp",
      "FromPort": 3300,
      "ToPort": 3300,
      "IpRanges": [{"CidrIp": "'"$HOSTED_VPC_CIDR"'"}]
    },
    {
      "IpProtocol": "tcp",
      "FromPort": 6789,
      "ToPort": 6789,
      "IpRanges": [{"CidrIp": "'"$HOSTED_VPC_CIDR"'"}]
    },
    {
      "IpProtocol": "tcp",
      "FromPort": 9283,
      "ToPort": 9283,
      "IpRanges": [{"CidrIp": "'"$HOSTED_VPC_CIDR"'"}]
    },
    {
      "IpProtocol": "tcp",
      "FromPort": 6800,
      "ToPort": 7300,
      "IpRanges": [{"CidrIp": "'"$HOSTED_VPC_CIDR"'"}]
    }
  ]' --output off
```

---

### Install DF on host cluster

Install DF operator from the operator hub.

Create storage cluster (make sure to enable host networking).

Create storage consumer and generate token.

Get endpoint from the storage cluster status.

---

### Install DF Client on hosted cluster

Install DF Client operator from the operator hub.

Create the storage client and configure it with the token and endpoint.

Verify storage client is connected and able to consume storage.

---

### Cleanup

Delete peering connection

```bash
REGION=us-east-1
KUBECONFIG_HOSTED=kubeconfig.hosted
```

```bash
HOSTED_INSTANCE_ID=$(oc get nodes -l node-role.kubernetes.io/worker \
  -o jsonpath='{.items[0].spec.providerID}' \
  --kubeconfig "$KUBECONFIG_HOSTED" | xargs basename)

HOSTED_VPC_ID=$(aws ec2 describe-instances \
  --region "$REGION" \
  --instance-ids "$HOSTED_INSTANCE_ID" \
  --query 'Reservations[0].Instances[0].VpcId' \
  --output text)

PCX_ID=$(aws ec2 describe-vpc-peering-connections \
  --region "$REGION" \
  --filters "Name=requester-vpc-info.vpc-id,Values=$HOSTED_VPC_ID" \
  --query 'VpcPeeringConnections[].VpcPeeringConnectionId' \
  --output text)

aws ec2 delete-vpc-peering-connection \
  --region "$REGION" \
  --vpc-peering-connection-id "$PCX_ID"
```

Destroy hosted cluster

```bash
REGION=us-east-1
ROLE=nigoyal-hcp-cli-role
BASE_DOMAIN=ocs.syseng.devcluster.openshift.com
```

```bash
aws sts get-session-token --output json > sts-creds.json

ARN=$(aws iam get-role --role-name "$ROLE" --query 'Role.Arn' --output text)

hcp destroy cluster aws \
  --name hosted \
  --infra-id hosted \
  --namespace hosted \
  --region "$REGION" \
  --base-domain "$BASE_DOMAIN" \
  --sts-creds ./sts-creds.json \
  --role-arn "$ARN"
```

Delete aws iam role

```bash
ROLE=nigoyal-hcp-cli-role
```

```bash
aws iam delete-role-policy \
  --role-name "$ROLE" \
  --policy-name "$ROLE"-policy

aws iam delete-role \
  --role-name "$ROLE"
```

Delete s3 bucket

```bash
REGION=us-east-1
BUCKET=nigoyal-hcp-oidc-bucket
```

```bash
aws s3 rb s3://"$BUCKET" \
  --region "$REGION" \
  --force
```

---
