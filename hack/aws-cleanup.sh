#!/bin/bash

set -e

# Calculate age in hours from given time
calculate_age() {
    local launch_time=$1

    # Get the current time in ISO 8601 format
    current_time=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

    # Calculate the difference in seconds
    age_seconds=$(($(date -d "$current_time" +%s) - $(date -d "$launch_time" +%s)))

    # Convert seconds to hours
    age_hours=$((age_seconds / 3600))

    # Return the age in hours
    echo "$age_hours"
}


# Delete the instance older than 3 hours
delete_ec2_instance() {
    local region=$1
    local instance_id=$2

    # Get the launch time of the instance
    launch_time=$(aws ec2 describe-instances --region "$region" --instance-ids "$instance_id" \
        --query "Reservations[*].Instances[*].{LaunchTime:LaunchTime}" --output text)

    # If launch_time is empty return
    [[ -z "$launch_time" ]] && return

    age_hours=$(calculate_age "$launch_time")

    # Delete the instance if older than 3 hours
    if [ "$age_hours" -gt 3 ]; then
        echo "Deleting the instance $instance_id in region $region with launch time $launch_time"
        aws ec2 terminate-instances --region "$region" --instance-ids "$instance_id"
    fi
}


# Delete the volume older than 3 hours
delete_ec2_volume() {
    local region=$1
    local volume_id=$2

    # Get the create time of the volume
    create_time=$(aws ec2 describe-volumes --region "$region" --volume-ids "$volume_id" \
        --query "Volumes[*].{CreateTime:CreateTime}" --output text)

    # If create_time is empty return
    [[ -z "$create_time" ]] && return

    age_hours=$(calculate_age "$create_time")

    # Delete the volume if older than 3 hours
    if [ "$age_hours" -gt 3 ]; then
        echo "Deleting the volume $volume_id in region $region with create time $create_time"
        aws ec2 delete-volume --region "$region" --volume-id "$volume_id"
    fi
}


# Delete the nat gateway older than 3 hours
delete_ec2_nat_gateway() {
    local region=$1
    local nat_gateway_id=$2

    # Get the create time of the nat gateway
    create_time=$(aws ec2 describe-nat-gateways --region "$region" --nat-gateway-ids "$nat_gateway_id" \
        --query "NatGateways[*].{CreateTime:CreateTime}" --output text)

    # If create_time is empty return
    [[ -z "$create_time" ]] && return

    age_hours=$(calculate_age "$create_time")

    # Delete the nat gateway if older than 3 hours
    if [ "$age_hours" -gt 3 ]; then
        echo "Deleting the nat gateway $nat_gateway_id in region $region with create time $create_time"
        aws ec2 delete-nat-gateway --region "$region" --nat-gateway-id "$nat_gateway_id"
    fi
}


# Delete the load balancer older than 3 hours
delete_elb_load_balancer() {
    local region=$1
    local load_balancer_arn=$2

    # Get the created time of the load balancer
    created_time=$(aws elbv2 describe-load-balancers --region "$region" --load-balancer-arn "$load_balancer_arn" \
        --query "LoadBalancers[*].{CreatedTime:CreatedTime}" --output text)

    # If created_time is empty return
    [[ -z "$created_time" ]] && return

    age_hours=$(calculate_age "$created_time")

    # Delete the load balancer if older than 3 hours
    if [ "$age_hours" -gt 3 ]; then
        echo "Deleting the load balancer $load_balancer_arn in region $region with created time $created_time"
        aws elbv2 delete-load-balancer --region "$region" --load-balancer-arn "$load_balancer_arn"
    fi
}


for region in us-east-1 us-east-2 us-west-1 us-west-2; do
    # List ec2 instances which are running
    for instance_id in $(aws ec2 describe-instances --region "$region" \
        --query "Reservations[*].Instances[?State.Name=='running'].InstanceId" --output text); do

        delete_ec2_instance "$region" "$instance_id"
    done

    # List ec2 volumes which are available
    for volume_id in $(aws ec2 describe-volumes --region "$region" \
        --query "Volumes[?State=='available'].VolumeId" --output text); do

        delete_ec2_volume "$region" "$volume_id"
    done

    # List elb load balancer
    for load_balancer_arn in $(aws elbv2 describe-load-balancers --region "$region" \
        --query "LoadBalancers[*].LoadBalancerArn" --output text); do

        delete_elb_load_balancer "$region" "$load_balancer_arn"
    done

    # List ec2 nat gateways which are available
    for nat_gateway_id in $(aws ec2 describe-nat-gateways --region "$region" \
        --query "NatGateways[?State=='available'].NatGatewayId" --output text); do

        delete_ec2_nat_gateway "$region" "$nat_gateway_id"
    done

    # List ec2 network interfaces which are available
    for network_interface_id in $(aws ec2 describe-network-interfaces --region "$region" \
        --query "NetworkInterfaces[?Status=='available'].NetworkInterfaceId" --output text); do

        # Delete the ec2 network interfaces which are available
        echo "Deleting the network interface $network_interface_id in region $region"
        aws ec2 delete-network-interface --region "$region" --network-interface-id "$network_interface_id"
    done

    # TODO implement VPC deletion as per https://docs.aws.amazon.com/vpc/latest/userguide/delete-vpc.html
    # List ec2 vpc which are available except default
    #for vpc_id in $(aws ec2 describe-vpcs --region "$region" \
    #    --query "Vpcs[?State=='available'].VpcId" --filters "Name=isDefault,Values=false" --output text); do

    #    # Delete the ec2 vpc which are available except default
    #    echo "Deleting the vpc $vpc_id in region $region"
    #    aws ec2 delete-vpc --region "$region" --vpc-id "$vpc_id"
    #done

done
