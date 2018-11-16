# AppEngine Flex with OpenCensus Example

## Setup Credentials
1. Create a Google Cloud Platform Project
1. [Create a service account][CreateServiceAccountLink] with Trace Append permission. Furnish a new
JSON key and then set the credentials using the `GOOGLE_APPLICATION_CREDENTIALS` environment
variable or [using GCP Starter Core properties][GcpStarterCorePropertiesLink]. Alternatively, if you
have the [Google Cloud SDK][GoogleCloudSdkLink] installed and initialized and are logged in with
[application default credentials][ApplicationDefaultCredentialsLink], you can skip this step since
the sample will auto-discover those settings for you.
1. Enable the [Stackdriver Trace API][StackdriverTraceApiLink]

## Setup AppEngine Flex for Java
Follow the instructions in the [Quickstart for Java][AppEngineFlexLink] to setup AppEngine Flex.
You'll need to reference your project id to run the application.

## Setup CloudStorage
Follow the instructions in the [getting started][CloudStorageLink] to setup Cloud Storage. You'll
need to reference your project id to run the application.

## Running locally
* In one of the directory of this `README.md` file:

    mvn clean jetty:run-exploded

* Go to `http://localhost:8080/` to check that the Frontend is up.
* Go to `http://localhost:8080/init` if this is the first time you are using the Sample.
* Go to `http://localhost:8080/work` this can be called multiple times to generate traces.
* Go to `http://localhost:8080/cleanup` when you are done using the Sample. 

To see the traces, navigate to Stackdriver Trace console's [Trace List][TraceListLink] view.

## Deploy to the App Engine flexible environment

    mvn clean appengine:deploy

* Visit `http://PROJECTID.appspot.com`.

[AppEngineFlexLink]: https://cloud.google.com/appengine/docs/flexible/java/quickstart
[ApplicationDefaultCredentialsLink]: https://developers.google.com/identity/protocols/application-default-credentials
[CloudStorageLink]: https://github.com/GoogleCloudPlatform/google-cloud-java/tree/master/google-cloud-clients/google-cloud-storage#getting-started
[CreateServiceAccountLink]: https://cloud.google.com/docs/authentication/getting-started#creating_the_service_account
[GoogleCloudSdkLink]: https://cloud.google.com/sdk/
[TraceListLink]: https://console.cloud.google.com/traces/traces
[StackdriverTraceApiLink]: https://console.cloud.google.com/apis/api/cloudtrace.googleapis.com/overview
