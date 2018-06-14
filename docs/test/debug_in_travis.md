# Introduction

PouchContainer uses [Travis](https://travis-ci.com/) as the Continuous Integration tool. Sometimes, tests may pass on local machine while fail on Travis environment. Debug in Travis is provided.
This document will give a simple instruction about debugging in Travis.

## Turn on debug feature

For public repository, debug feature is off by default. An email to `support@travis-ci.com
` is needed to ask the Travis staff to turn on debug feature for your public repository.

## A real example

There is a [guide]( https://docs.travis-ci.com/user/running-build-in-debug-mode/) about how to use debug mode, you should read it before starting. After that, let's give a real example.

Take a Travis build from my repository as example:
`https://travis-ci.org/Letty5411/pouch/builds/308762622?utm_source=github_status&utm_medium=notification`

### Command to start debug mode

From the guide provided by Travis, we know that a `curl` request is needed to be constructed to start the debug mode, the command template is as following:

```shell
#curl -s -X POST \
   -H "Content-Type: application/json" \
   -H "Accept: application/json" \
   -H "Travis-API-Version: 3" \
   -H "Authorization: token <TOKEN>" \
   -d '{ "quiet": true }' \
   https://api.travis-ci.org/job/{ID}/debug
```

For the command, the `TOKEN` and `job ID` is needed. How to get them?

#### Get token

Here is the [guide](https://docs.travis-ci.com/user/triggering-builds/) of how to generate token. It only needs three steps:

1.Install Travis CI command line client according the this [guide](https://github.com/travis-ci/travis.rb).

```
# gem install travis -v 1.8.8 --no-rdoc --no-ri

# ruby -v
ruby 2.0.0p598 (2014-11-13) [x86_64-linux]

# travis  -v
1.8.8

```

2.Log into Travis, need your GitHub username and password and will get you authenticated.

```
#travis login --org
We need your GitHub login to identify you.
This information will not be sent to Travis CI, only to api.github.com.
The password will not be displayed.

Try running with --github-token or --auto if you don't want to enter your password anyway.

Username: xxxx
Password for xxxx: ****
Successfully logged in as xxxx!
```

3.Get an API token using the Travis CI command line client, if everything goes well you will get your token.

```
#travis token --org

Your access token is yourtoken
```

#### Get job ID

In my case, the job ID for this build is: 308880940, which you can find in the build system information section of the log.

```
Build system information
Build language: go
Build group: stable
Build dist: trusty
Build id: 308880939
Job id: 308880940
```

### Start debug mode

Now both token and job ID is ready, a curl command could be run to start debug mode:

```
#curl -s -X POST   -H "Content-Type: application/json"   -H "Accept: application/json"   -H "Travis-API-Version: 3"   -H "Authorization: token yourtoken" -d "{\"quiet\": true}"  https://api.travis-ci.org/job/308880940/debug
{
  "@type": "pending",
  "job": {
    "@type": "job",
    "@href": "/job/308880940",
    "@representation": "minimal",
    "id": 308880940
  },
  "state_change": "created",
  "resource_type": "job"
}
```

Head back to the web UI, wait a few minute. Then in the log, we will see the following output indicating the VM has been connected:

```
Debug build initiated by Letty5411
Setting up debug tools.
Preparing debug sessions.
Use the following SSH command to access the interactive debugging environment:
ssh gpXeOCVaz6bPsHxrhplC41ITu@to2.tmate.io
This build is running in quiet mode. No session output will be displayed.
This debug build will stay alive for 30 minutes.
....

```

Setting up debug tools.
Preparing debug sessions.
Use the following SSH command to access the interactive debugging environment:

```
ssh ukjiuCEkxBBnRAe32Y8xCH0zj@ny2.tmate.io
```

### SSH the VM and enjoy debugging

Now you could debug on the Travis VM through SSH command from your computer. Once you're done, just type exit and your build will terminate.

```
Run individual commands; or execute configured build phases
with `travis_run_*` functions (e.g., `travis_run_before_install`).

For more information, consult https://docs.travis-ci.com/user/running-build-in-debug-mode/, or email support@travis-ci.com.

travis@travis-job-885b6c7c-5885-469e-90bb-d3014cc72000:~/gopath/src/github.com/alibaba/pouch$
travis@travis-job-885b6c7c-5885-469e-90bb-d3014cc72000:~/gopath/src/github.com/alibaba/pouch$ pwd
/home/travis/gopath/src/github.com/alibaba/pouch
travis@travis-job-885b6c7c-5885-469e-90bb-d3014cc72000:~/gopath/src/github.com/alibaba/pouch$ whoami
travis
travis@travis-job-885b6c7c-5885-469e-90bb-d3014cc72000:~/gopath/src/github.com/alibaba/pouch$ uname -r
4.4.0-93-generic
travis@travis-job-885b6c7c-5885-469e-90bb-d3014cc72000:~/gopath/src/github.com/alibaba/pouch$ uname -a
Linux travis-job-885b6c7c-5885-469e-90bb-d3014cc72000 4.4.0-93-generic #116~14.04.1-Ubuntu SMP Mon Aug 14 16:07:05 UTC 2017 x86_64 x86_64 x86_64 GNU/Linux
travis@travis-job-885b6c7c-5885-469e-90bb-d3014cc72000:~/gopath/src/github.com/alibaba/pouch$
```

Finally, once in the SSH session, these [bash functions](https://docs.travis-ci.com/user/running-build-in-debug-mode/#Things-to-do-once-you-are-inside-the-debug-VM) will come in handy to run the different phases in your build.

## Acknowledgements

A lot thanks should give to Carla Iriberri, staff from Travis, who gives me this specific instruction. This doc is extracted from the email between Carla Iriberri and me.
By the way, if you have any question about Travis, please do not hesitate to write email to
`support@travis-ci.com`, they have really quick response and good support. Thanks again for Carla Iriberri's help.
