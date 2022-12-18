## Phaction Architecture

1. Upload MOV video files thru UI

2. Makulu service receives video, sends it to an S3 bucket, constructs and sends a message containing video S3 bucket/key and
    
    metadata (file type) to kafka topic 'upload-mov'

3. Chyme Tasker is listening for messages on 'upload-mov' topic, creates a task and sends as a message to Kafka topic 

    'process-task'

4. Chyme Worker listening for messages on 'process-task' starts a new container with the appropriate image for execution,

    places the resulting Dash manifest in an S3 bucket, metadata (including S3 bucket/key) in D-Graph and geometries in 
    
    GeoMesa/GeoServer.

5. Frontend can query using geometry (OpenLayers), date, activity type