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

    # If launch_time is empty return
    [[ -z "$create_time" ]] && return

    age_hours=$(calculate_age "$create_time")

    # Delete the volume if older than 3 hours
    if [ "$age_hours" -gt 3 ]; then
        echo "Deleting the volume $volume_id in region $region with create time $create_time"
        aws ec2 delete-volume --region "$region" --volume-id "$volume_id"
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
        --query "Volumes[?State=='available'].{VolumeId:VolumeId}" --output text); do

        delete_ec2_volume "$region" "$volume_id"
    done
done
