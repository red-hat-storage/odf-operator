/*
Copyright 2021 Red Hat OpenShift Data Foundation.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import "strings"

var AllQuickStarts = [][]byte{[]byte(gettingStartedQS), []byte(odfConfigAndManagementQS)}

var gettingStartedQS = strings.ReplaceAll(`
apiVersion: console.openshift.io/v1
kind: ConsoleQuickStart
metadata:
  name: "getting-started-odf"
spec:
  displayName: Getting started with OpenShift Data Foundation
  durationMinutes: 10
  icon: data:image/svg+xml;base64,PHN2ZyBlbmFibGUtYmFja2dyb3VuZD0ibmV3IDAgMCAxMDAgMTAwIiBoZWlnaHQ9IjEwMCIgdmlld0JveD0iMCAwIDEwMCAxMDAiIHdpZHRoPSIxMDAiIHhtbG5zPSJodHRwOi8vd3d3LnczLm9yZy8yMDAwL3N2ZyI+PHBhdGggZD0ibTY2LjcgNTUuOGM2LjYgMCAxNi4xLTEuNCAxNi4xLTkuMiAwLS42IDAtMS4yLS4yLTEuOGwtMy45LTE3Yy0uOS0zLjctMS43LTUuNC04LjMtOC43LTUuMS0yLjYtMTYuMi02LjktMTkuNS02LjktMy4xIDAtNCA0LTcuNiA0LTMuNSAwLTYuMS0yLjktOS40LTIuOS0zLjIgMC01LjIgMi4xLTYuOCA2LjYgMCAwLTQuNCAxMi41LTUgMTQuMy0uMS4zLS4xLjctLjEgMSAuMSA0LjcgMTkuMiAyMC42IDQ0LjcgMjAuNm0xNy4xLTZjLjkgNC4zLjkgNC44LjkgNS4zIDAgNy40LTguMyAxMS40LTE5LjEgMTEuNC0yNC42IDAtNDYuMS0xNC40LTQ2LjEtMjMuOSAwLTEuMy4zLTIuNi44LTMuOS04LjkuNS0yMC4zIDIuMS0yMC4zIDEyLjIgMCAxNi41IDM5LjIgMzYuOSA3MC4yIDM2LjkgMjMuOCAwIDI5LjgtMTAuNyAyOS44LTE5LjIgMC02LjctNS44LTE0LjMtMTYuMi0xOC44IiBmaWxsPSIjZWQxYzI0Ii8+PHBhdGggZD0ibTgzLjggNDkuOGMuOSA0LjMuOSA0LjguOSA1LjMgMCA3LjQtOC4zIDExLjQtMTkuMSAxMS40LTI0LjYgMC00Ni4xLTE0LjQtNDYuMS0yMy45IDAtMS4zLjMtMi42LjgtMy45bDEuOS00LjhjLS4xLjMtLjEuNy0uMSAxIDAgNC44IDE5LjEgMjAuNyA0NC43IDIwLjcgNi42IDAgMTYuMS0xLjQgMTYuMS05LjIgMC0uNiAwLTEuMi0uMi0xLjh6IiBmaWxsPSIjMDEwMTAxIi8+PC9zdmc+
  description: Learn how to create persistent files, object storage and connect it with your applications.
  introduction: >-
    RedHat OpenShift Data Foundation provides a highly integrated collection of cloud storage and data services for OpenShift Container Platform.

    In this tour, you'll learn how to add block, file, or object storage to your applications, and how to monitor your storage resources in the OpenShift Data Foundation dashboards.

      1. Connecting applications to block or file storage (persistent volumes)

      1. Connecting applications to object storage (object buckets)

      1. Using the dashboards to monitor OpenShift Data Foundation resources

  tasks:
    -
      title: Connecting applications to block or file storage (persistent volumes)
      description: >-

        If you want yours data to live longer than yours pods, you need persistent volumes.

        Persistent volumes exist outside the pod lifecycle, meaning that your data is retained even after your pod has been restarted, rescheduled, or deleted. You might use persistent volumes with a MySQL or WordPress application to store information for longer than the lifetime of any individual application pod.


        After an administrator has set up an OpenShift Data Foundation cluster, developers can use persistent volume claims (PVCs) to request persistent volume (PV) resources without needing to know anything specific about the underlying storage infrastructure.


        **Connect your application with a PVC**

           1. Click Workloads > Deployments in the navigation menu on the left.

           2. Select your project from the Project drop-down and find your application in the list of deployments.

           3. Click the Action menu &*⋮&* > Add Storage

           4. Select &*Create new claim&*, You would select &*Use existing claim&* here if you wanted to restore an existing persistent volume to a redeployed application.

           5. Select the appropriate type of storage for your application.

              - &*For block storage&*, select ocs-storagecluster-ceph-rbd.

              - &*For file storage&*, select ocs-storagecluster-cephfs.


           6. Specify details of the storage you want to create.

           7. Click Save.
      review:
        instructions: >-
          To verify that your application is using persistent volume claim:

            - Click the name of the deployment that you assigned storage to.

            - On the deployment details page, look at Type in the Volumes section to verify the type of the PVC you attached.

            - Click the PVC name and verify the storage class name in the PersistentVolumeClaim Overview page.

        failedTaskHelp: Try the steps again.
    -
      title: Connecting application to Object Storage
      description: >-

        Object Buckets provide an easy way to consume object storage across OpenShift Data Foundation. Use your object service endpoint, access key, and secret key to add your object service provider to OpenShift Data Foundation as a backing store. See Adding storage resources for hybrid or multicloud docs.

        After your object service is configured, you can create an Object Bucket Claim and connect it to your application.

          1. Click Storage > Object Bucket Claims.

          2. Click Create Object Bucket Claim.

          3. Specify a name for your claim and select an appropriate storage class for your application:

             - &*To use on-premises object storage&* select ocs-storagecluster-ceph-rgw.

             - &*To use Multicloud Object Gateway&* select openshift-storage.noobaa.io.

          4. Click Create.

          5. Click the Action menu &*⋮&* > Attach to deployment.

          6. Select the application to attach the Object Bucket Claim to and click Attach.

      review:
        instructions: >-
          To verify that your application is using an object bucket claim:

            - Click the name of the deployment that you assigned storage to.

            - On the deployment click on Environment tab and check if a the new secret and config map were added.

        failedTaskHelp: Try the steps again.
    -
      title: Using the dashboards to monitor OpenShift Data Foundation resources
      description:  >-
        You can monitor any storage resource manage by Openshift Container Storage on Openshift Data Foundation views:


        Click Storage > Openshift Data Foundation

        1. The ODF overview gives you a high level view for all storage systems.

        1. To get deeper view for a specific system, you can drill down to see system overview.

            - The Block & File dashboard tab shows the state of the Openshift Container Storage as whole, as well a the state of the persistent volumes.

            - The Object dashboard shows the state of the Multicloud Object Gateway, RADOS Object Gateway, and any object claims.

  conclusion: You finished the Getting Started Quickstart
`, "&*", "`")

const odfConfigAndManagementQS = `
apiVersion: console.openshift.io/v1
kind: ConsoleQuickStart
metadata:
  name: odf-configuration
spec:
  displayName: OpenShift Data Foundation Configuration & Management
  durationMinutes: 5
  icon: data:image/svg+xml;base64,PHN2ZyBlbmFibGUtYmFja2dyb3VuZD0ibmV3IDAgMCAxMDAgMTAwIiBoZWlnaHQ9IjEwMCIgdmlld0JveD0iMCAwIDEwMCAxMDAiIHdpZHRoPSIxMDAiIHhtbG5zPSJodHRwOi8vd3d3LnczLm9yZy8yMDAwL3N2ZyI+PHBhdGggZD0ibTY2LjcgNTUuOGM2LjYgMCAxNi4xLTEuNCAxNi4xLTkuMiAwLS42IDAtMS4yLS4yLTEuOGwtMy45LTE3Yy0uOS0zLjctMS43LTUuNC04LjMtOC43LTUuMS0yLjYtMTYuMi02LjktMTkuNS02LjktMy4xIDAtNCA0LTcuNiA0LTMuNSAwLTYuMS0yLjktOS40LTIuOS0zLjIgMC01LjIgMi4xLTYuOCA2LjYgMCAwLTQuNCAxMi41LTUgMTQuMy0uMS4zLS4xLjctLjEgMSAuMSA0LjcgMTkuMiAyMC42IDQ0LjcgMjAuNm0xNy4xLTZjLjkgNC4zLjkgNC44LjkgNS4zIDAgNy40LTguMyAxMS40LTE5LjEgMTEuNC0yNC42IDAtNDYuMS0xNC40LTQ2LjEtMjMuOSAwLTEuMy4zLTIuNi44LTMuOS04LjkuNS0yMC4zIDIuMS0yMC4zIDEyLjIgMCAxNi41IDM5LjIgMzYuOSA3MC4yIDM2LjkgMjMuOCAwIDI5LjgtMTAuNyAyOS44LTE5LjIgMC02LjctNS44LTE0LjMtMTYuMi0xOC44IiBmaWxsPSIjZWQxYzI0Ii8+PHBhdGggZD0ibTgzLjggNDkuOGMuOSA0LjMuOSA0LjguOSA1LjMgMCA3LjQtOC4zIDExLjQtMTkuMSAxMS40LTI0LjYgMC00Ni4xLTE0LjQtNDYuMS0yMy45IDAtMS4zLjMtMi42LjgtMy45bDEuOS00LjhjLS4xLjMtLjEuNy0uMSAxIDAgNC44IDE5LjEgMjAuNyA0NC43IDIwLjcgNi42IDAgMTYuMS0xLjQgMTYuMS05LjIgMC0uNiAwLTEuMi0uMi0xLjh6IiBmaWxsPSIjMDEwMTAxIi8+PC9zdmc+
  description: Learn how to configure OpenShift Data Foundation to meet your deployment
    needs.
  prerequisites: ["Getting Started with OpenShift Data Foundation", "Install the Openshift Data Foundation" ]
  introduction: In this tour, you will learn about the various configurations available
    to customize your OpenShift® Data Foundation deployment.
  tasks:
    - title: Expand the ODF Storage System
      description: |-
        When we install the ODF operator we created a storage cluster, chose
        the cluster size, provisioned the underlying storage subsystem, deployed necessary
        drivers, and created the storage classes to allow the OpenShift users to easily
        provision and consume storage services that have just been deployed

        When the capacity of the cluster is about to runout we will notify you.

        **To expand the OCS storage cluster follow these steps:**
        1. Go to installed operators page and click on **OpenShift Data Foundation**
        2. Go to storage cluster tab
        3. Click on the **3 dots icon**
        4. Click on add capacity
        5. Use the expand cluster modal if your capacity is about to runout.
      review:
        instructions: |-
          ####  To verify that you have expanded your storage cluster.
          Did you expand your cluster?
        failedTaskHelp: This task isn’t verified yet. Try the task again.
      summary:
        success: You have expanded the Storage Cluster for the ODF operator!
        failed: Try the steps again.
    - title: Bucket Class Configuration
      description: |-

          Bucket class policy determines the bucket's data location. Its set of policies which apply to all buckets (OBCs) created with the specific bucket class. These policies include: placement, namespace, caching

          There are two types of Bucket Classes:
           - **Standard:** Data will be ingested by Multi Cloud Object Gateway, deduped, compressed and encrypted.
           - **Namespace:** Data will be stored as is (no dedup, compression, encryption) on the namespace stores.

          **Create a new Bucket class**

          1. Go to installed operators page and click on OpenShift Data Foundation,
          2. Go to bucket class tab.
          3. Click on **Create Bucket Class**
          4. Follow the wizard steps to  finish creation process.
      review:
        instructions: |-
          ####  To verify that you have created bucket class and backing store.
          Is the Bucket Class in ready state?
        failedTaskHelp: This task isn’t verified yet. Try the task again.
      summary:
        success: You have successfully created bucket class
        failed: Try the steps again.
  conclusion: Congrats, the OpenShift Data Foundation operator is ready to use.
`
