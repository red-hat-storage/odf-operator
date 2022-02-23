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
      title: Connect applications to block or file storage (PersistentVolumeClaims)
      description: >-

        PersistentVolumes (PVs) allow your data to exist beyond the pod's lifecycle, even after you restart, reschedule, or delete it.


        After an administrator sets up an OpenShift Data Foundation StorageSystem, developers can use PersistentVolumeClaims (PVCs) to request PV resources without needing to know anything specific about their underlying storage infrastructure.


        **Connect your application with a PVC:**

           1. In the side navigation menu, select **Workloads > Deployments**.

           2. Select your project from the **Project** dropdown and find your application in the list of deployments.

           3. Open the action menu &*⋮&* and select **Add storage**.

           4. Select **Storage type** PersistentVolumeClaim if it isn’t already selected.

           5. To restore an existing PV to a redeployed application, select **Use existing claim**. To create a claim, select **Create new claim**.

           6. Select the appropriate storage type for your application:

              - &*Block storage:&* Select **ocs-storagecluster-ceph-rbd**.

              - &*File storage:&* Select **ocs-storagecluster-cephfs**.

           7. Specify your storage details. If you’re using an existing claim, specify your mount path details. 

           8. Select **Save**.
      review:
        instructions: >-
          To verify that your application is using PersistentVolumeClaim:

            - Select the name of the Deployment that you assigned storage to.

            - On the **Deployment details** page, look at Type in the Volumes section to verify the type of the PVC you attached.

            - Select the PVC name and verify the StorageClass name in the PersistentVolumeClaim Details page.

        failedTaskHelp: Try the steps again.
    -
      title: Connect applications to object storage (Object Bucket Claims)
      description: >-

        Object Bucket Claims provide an easy way to consume object storage across OpenShift Data Foundation. 
        
        
        Use your object service endpoint, access key, and secret key to add your object service provider to OpenShift Data Foundation as a BackingStore.

       
        **To create an Object Bucket Claim and connect it to your application:**

          1. Select **Storage > Object Bucket Claims**.

          2. Select **Create Object Bucket Claim**.

          3. Enter a name for your claim and select an appropriate StorageClass for your application:

             - &*On-premises object storage:&* Select **ocs-storagecluster-ceph-rgw**.

             - &*Multicloud Object Gateway:&* Select **openshift-storage.noobaa.io**.

          4. Select Object Bucket Claim for StorageClass **openshift-storage.noobaa.io**.
          
          5. To create your Object Bucket Claim, select **Create**.

          6. On the **Object Bucket Claims** page, verify that your object bucket claim’s status is **Bound**.

          7. Open the action menu &*⋮&* and select **Attach to deployment**.

          8. To attach your object bucket claim to your application, navigate to the pop-up that appears on your screen and select its name from the dropdown list under **Deployment Name**.

          9. Select **Attach**.

      review:
        instructions: >-
          To verify that your application is using an Object Bucket Claim:

            - Select the name of the Deployment that you assigned storage to.

            - On the **Deployment details** page, select the **Environment** tab and check if a new **Secret** and **ConfigMap** were added.

        failedTaskHelp: Try the steps again.
    -
      title: Use the dashboards to monitor OpenShift Data Foundation resources
      description:  >-

        Monitor any storage resource managed by Openshift Data Foundation through various dashboard overviews.


        In the side navigation, select **Storage > Openshift Data Foundation** to access the Openshift Data Foundation dashboard view.

        1. Observe high level insights for all your StorageSystems with the overview screen.

        1. To access more specific information for a system, select its **System Capacity** tile and drill down to its system overview.

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
  displayName: Configure and manage OpenShift Data Foundation
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
        
        - Created StorageClasses.
        

        Now, you’re ready to easily provision and consume your deployed storage services.


        Monitor your storage regularly so that you don't run out of storage space.


        As you consume storage, you'll receive cluster capacity alerts at 75% (near-full) capacity and 85% (full) capacity. Always address capacity warnings promptly.


        **To expand your StorageSystem:**

        1. In the navigation menu, select **Storage > OpenShift Data Foundation**.

        2. Navigate to the **Storage Systems** tab.

        3. Open the action menu &*⋮&*.

        4. Select **Add Capacity**.

        5. Select your desired StorageClass from the dropdown.

        6. Select **Add**. Once your selected StorageSystem’s status changes to **Ready**, you’ve successfully expanded your StorageSystem.

      review:
        instructions: |-
          ####  To verify that you expanded your StorageSystem:
          Navigate to the **Overview > Block and File** for this StorageSystem. Under the **Raw capacity** section, has your **Available** capacity increased ?
        failedTaskHelp: This task isn’t verified yet. Try the task again.
      summary:
        success: You have expanded the StorageSystem for the ODF operator!
        failed: Try the steps again.
    - title: Configure BucketClass
      description: |-

          BucketClass determines a bucket's data location and provides a set of policies (placement, namespace, caching) that applies to all buckets created with the same class.


          BucketClasses occur in two types:

           - &*Standard:&* Data is ingested by Multicloud Object Gateway, deduped, compressed and encrypted.
           - &*Namespace:&* Data is stored as-is on the NamespaceStores without being deduped, compressed or encrypted.


          **To create a BucketClass:**

          1. In the main navigation menu, select **Storage > OpenShift Data Foundation**.

          2. Select **Bucket Class** tab.

          3. Select **Create Bucket Class**

          4. In the wizard, follow each step to create your BucketClass.
      review:
        instructions: |-
          ####  To verify that you created BucketClass and BackingStore:
          Is the BucketClass in **Ready** state?
        failedTaskHelp: This task isn’t verified yet. Try the task again.
      summary:
        success: You have successfully created BucketClass
        failed: Try the steps again.

  conclusion: You're ready to go! Now you can customize your StorageSystems in OpenShift Data Foundation.
`, "&*", "`")
