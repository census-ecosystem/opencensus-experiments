
mkdir .javadep
cd .javadep
git clone https://github.com/census-instrumentation/opencensus-java.git
cd opencensus-java
VERSION=`egrep CURRENT_OPENCENSUS_VERSION build.gradle | awk '{print $3}' | sed 's/"//g'`
export $VERSION


