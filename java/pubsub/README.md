# Using OpenCensus and Stackdriver with Cloud Pub/Sub and the Google Cloud Client libraries

<a href="https://console.cloud.google.com/cloudshell/open?git_repo=https://github.com/GoogleCloudPlatform/java-docs-samples&page=editor&open_in_editor=pubsub/cloud-client/README.md">
<img alt="Open in Cloud Shell" src ="http://gstatic.com/cloudssh/images/open-btn.png"></a>

[Google Cloud Pub/Sub][pubsub] is a fully-managed real-time messaging service that allows you to
send and receive messages between independent applications.
This sample Java application demonstrates how to use [OpenCensus][opencensus] with the Pub/Sub API using
the [Google Cloud Client Library for Java][google-cloud-java].

[pubsub]: https://cloud.google.com/pubsub/
[opencensus]: https://opencensus.io
[google-cloud-java]: https://github.com/GoogleCloudPlatform/google-cloud-java
[google-cloud-examples]: https://github.com/GoogleCloudPlatform/google-cloud-java

This example is based on the Pub/Sub example in the 
[Google Cloud Platform Java examples](https://github.com/GoogleCloudPlatform/java-docs-samples).
For more samples, see the samples in 
[google-cloud-java](https://github.com/GoogleCloudPlatform/google-cloud-java/tree/master/google-cloud-examples/src/main/java/com/google/cloud/examples/pubsub).

## Quickstart

#### Setup
- Install [Maven](http://maven.apache.org/).
- [Enable](https://console.cloud.google.com/apis/api/pubsub.googleapis.com/overview) Pub/Sub API.
- Set up [authentication](https://cloud.google.com/docs/authentication/getting-started).

#### Build
- Build your project with:
```
  mvn clean package
```

#### Create a new topic
```
  mvn exec:java -Dexec.mainClass=com.example.pubsub.CreateTopicExample -Dexec.args=my-topic
```

#### Create a subscription
```
  mvn exec:java -Dexec.mainClass=com.example.pubsub.CreatePullSubscriptionExample -Dexec.args="my-topic my-sub"
```

#### Publish messages
```
  mvn exec:java -Dexec.mainClass=com.example.pubsub.PublisherExample -Dexec.args="my-topic 3"
```
Publishes 3 messages to the topic `my-topic`.

#### Receive messages
```
   mvn exec:java -Dexec.mainClass=com.example.pubsub.SubscriberExample -Dexec.args=my-sub
```
Subscriber will continue to listen on the topic for 5 minutes and print out message id and data as messages are received.


