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
  durationMinutes: 5
  icon: data:image/svg+xml;base64,PHN2ZyBlbmFibGUtYmFja2dyb3VuZD0ibmV3IDAgMCAxMDAgMTAwIiBoZWlnaHQ9IjEwMCIgdmlld0JveD0iMCAwIDEwMCAxMDAiIHdpZHRoPSIxMDAiIHhtbG5zPSJodHRwOi8vd3d3LnczLm9yZy8yMDAwL3N2ZyI+PHBhdGggZD0ibTY2LjcgNTUuOGM2LjYgMCAxNi4xLTEuNCAxNi4xLTkuMiAwLS42IDAtMS4yLS4yLTEuOGwtMy45LTE3Yy0uOS0zLjctMS43LTUuNC04LjMtOC43LTUuMS0yLjYtMTYuMi02LjktMTkuNS02LjktMy4xIDAtNCA0LTcuNiA0LTMuNSAwLTYuMS0yLjktOS40LTIuOS0zLjIgMC01LjIgMi4xLTYuOCA2LjYgMCAwLTQuNCAxMi41LTUgMTQuMy0uMS4zLS4xLjctLjEgMSAuMSA0LjcgMTkuMiAyMC42IDQ0LjcgMjAuNm0xNy4xLTZjLjkgNC4zLjkgNC44LjkgNS4zIDAgNy40LTguMyAxMS40LTE5LjEgMTEuNC0yNC42IDAtNDYuMS0xNC40LTQ2LjEtMjMuOSAwLTEuMy4zLTIuNi44LTMuOS04LjkuNS0yMC4zIDIuMS0yMC4zIDEyLjIgMCAxNi41IDM5LjIgMzYuOSA3MC4yIDM2LjkgMjMuOCAwIDI5LjgtMTAuNyAyOS44LTE5LjIgMC02LjctNS44LTE0LjMtMTYuMi0xOC44IiBmaWxsPSIjZWQxYzI0Ii8+PHBhdGggZD0ibTgzLjggNDkuOGMuOSA0LjMuOSA0LjguOSA1LjMgMCA3LjQtOC4zIDExLjQtMTkuMSAxMS40LTI0LjYgMC00Ni4xLTE0LjQtNDYuMS0yMy45IDAtMS4zLjMtMi42LjgtMy45bDEuOS00LjhjLS4xLjMtLjEuNy0uMSAxIDAgNC44IDE5LjEgMjAuNyA0NC43IDIwLjcgNi42IDAgMTYuMS0xLjQgMTYuMS05LjIgMC0uNiAwLTEuMi0uMi0xLjh6IiBmaWxsPSIjMDEwMTAxIi8+PC9zdmc+
  description: Learn how to create persistent files, object storage and connect it with your applications.
  introduction: >-
    **Red Hat OpenShift® Data Foundation** provides a highly integrated collection of cloud storage and data services for OpenShift Container Platform.


    In this tour, you'll learn how to:

      - Add block, file, or object storage to your applications.

      - Monitor your storage resources with OpenShift Data Foundation dashboards.

  tasks:
    -
      title: Connecting applications to block or file storage (PersistentVolumeClaims)
      description: >-

        PersistentVolumes (PVs) allow your data to exist beyond your pod's lifecycle, even after you restart, rescheduled, or delete it.


        After an administrator sets up an OpenShift Data Foundation StorageSystem, developers can use PersistentVolumeClaims (PVCs) to request PV resources without needing to know anything specific about their underlying storage infrastructure.


        **Connect your application with a PVC:**

           1. In the side navigation menu, select **Workloads > Deployments**.

           2. Select your project from the **Project** dropdown and find your application in the list of deployments.

           3. Open the action menu &*⋮&* and select **Add Storage**.

           4. To create a claim, select **Create new claim**. To restore an existing PV to a redeployed application, select **Use existing claim**.

           5. Select the appropriate storage type for your application:

              - &*Block storage:&* Select **ocs-storagecluster-ceph-rbd**.

              - &*File storage:&* Select **ocs-storagecluster-cephfs**.

           6. Specify storage details.

           7. Click **Save**.
      review:
        instructions: >-
          To verify that your application is using PersistentVolumeClaim:

            - Click the name of the deployment that you assigned storage to.

            - On the deployment details page, look at Type in the Volumes section to verify the type of the PVC you attached.

            - Click the PVC name and verify the storage class name in the PersistentVolumeClaim Overview page.

        failedTaskHelp: Try the steps again.
    -
      title: Connecting application to object storage (Object Bucket Claims)
      description: >-

        Object Bucket Claims provide an easy way to consume object storage across OpenShift Data Foundation. 
        
        
        Use your object service endpoint, access key, and secret key to add your object service provider to OpenShift Data Foundation as a BackingStore. See Learn how to add storage resources for hybrid or multicloud docs.

       
        **To create an Object Bucket Claim and connect it to your application:**

          1. Select **Storage > Object Bucket Claims**.

          2. Select **Create Object Bucket Claim**.

          3. Enter a name for your claim and select an appropriate StorageClass for your application:

             - &*On-premises object storage:&* Select **ocs-storagecluster-ceph-rgw**.

             - &*Multicloud Object Gateway:&* Select **openshift-storage.noobaa.io**.

          4. To create your Object Bucket Claim, select **Create**.

          5. Open the action menu &*⋮&* and select **Attach to deployment**.

          6. To attach the Object Bucket to your application, select its name from the application list.

      review:
        instructions: >-
          To verify that your application is using an Object Bucket Claim:

            - Click the name of the deployment that you assigned storage to.

            - On the deployment click on Environment tab and check if a the new secret and config map were added.

        failedTaskHelp: Try the steps again.
    -
      title: Using the dashboards to monitor OpenShift Data Foundation resources
      description:  >-

        Monitor any storage resource manage by Openshift Data Foundation through various dashboard overviews.


        In the side navigation, select **Storage > Openshift Data Foundation** to access the Openshift Data Foundation dashboard view.

        1. Observe high level insights for all your StorageSystems with the overview screen.

        1. To access more specific information for a system, drill down to its system overview.

            - The **Block & File overview** tab shows the holistic state of Openshift Data Foundation and the state of any PersistentVolumes.

            - The **Object overview** tab shows the state of the Multicloud Object Gateway, RADOS Object Gateway, and any Object Claims.

  conclusion: Now you're ready to add storage to your applications and monitor your storage resources.


  nextQuickStart:
  - "odf-configuration"
`, "&*", "`")

var odfConfigAndManagementQS = strings.ReplaceAll(`
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
  introduction: In this tour, you will learn how to customize your **Red Hat OpenShift® Data Foundation** StorageSystems.
  tasks:
    - title: Expand your StorageSystem
      description: |-
        When you installed the OpenShift Data Foundation operator, you:
        
        - Created a StorageSystem.
        
        - Set a cluster size. 
        
        - Provisioned a storage subsystem.
        
        - Deployed necessary drivers. 
        
        - Created StorageClasses 
        

        These installation actions enable you to easily provision and consume your deployed storage services.


        Monitor your storage regularly so that you don't run out of storage space.


        As you consume storage, you'll receive cluster capacity alerts at 75% capacity (near-full) and 85% (full) capacity. Always address capacity warnings promptly.


        **To expand your StorageSystem:**

        1. In the navigation menu, select **Storage > OpenShift Data Foundation**.

        2. Navigate to the **Storage Systems** tab.

        3. Open the action menu &*⋮&*.

        4. Select **Add Capacity**.

      review:
        instructions: |-
          ####  To verify that you have expanded your StorageSystem.
          Did you expand your StorageSystem?
        failedTaskHelp: This task isn’t verified yet. Try the task again.
      summary:
        success: You have expanded the StorageSystem for the ODF operator!
        failed: Try the steps again.
    - title: Configure BucketClass
      description: |-

          BucketClass determines a bucket's data location and provides a set of policies (placement, namespace, caching) that apply to all buckets created with the same class.


          BucketClasses occur in two types:

           - &*Standard:&* Data is ingested by Multicloud Object Gateway, deduped, compressed and encrypted.
           - &*Namespace:&* Data is stored as-is on the NamespaceStores without being deduped, compressed or encrypted.


          **To create a BucketClass:**

          1. in the navigation menu, select **Installed Operators > OpenShift Data Foundation**.

          2. Select **Bucket Class** tab.

          3. Select **Create Bucket Class**

          4. In the wizard, follow each step to create your BucketClass.
      review:
        instructions: |-
          ####  To verify that you have created BucketClass and BackingStore.
          Is the BucketClass in ready state?
        failedTaskHelp: This task isn’t verified yet. Try the task again.
      summary:
        success: You have successfully created BucketClass
        failed: Try the steps again.

  conclusion: You're ready to go! Now you can customize your StorageSystems in OpenShift Data Foundation.
`, "&*", "`")
