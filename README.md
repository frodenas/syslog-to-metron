# Syslog server to forward log messages to Metron

This is an experimental utility to forward syslog messages to [Cloud Foundry Metron](https://github.com/cloudfoundry/loggregator/tree/develop/src/metron).

## Disclaimer

This is a work in progress. It is suitable for experimentation and may not become supported in the future.

## Sample usage

### Deploy a sample application

Deploy a sample application to your Cloud Foundry environment and look for the application `guid`:

```
CF_TRACE=true cf app <APP-NAME> | grep guid
```

### Run the utility

Build and run this utility. [Metron](https://github.com/cloudfoundry/loggregator/tree/develop/src/metron) only listens to local network interfaces, so you must run the `syslog-to-metron` utility on the machine where the `metron_agent` process is running. You also need to specify the application `guid`.

```
syslog-to-metron -debug -metron-address 127.0.0.1:3457 -metron-origin my-service -syslog-address 0.0.0.0:10514 -syslog-protocol UDP -syslog-format RFC3164 -source-type SRV -source-instance 0 -app-id <APP-GUID>
```

### Forward your logs to the utility syslog server

Run a service and forward the logs the the `syslog-to-metron` syslog server. In this example we are using [Docker](https://www.docker.com/) and the [syslog logging driver](https://docs.docker.com/engine/admin/logging/overview/).

```
docker run --name redis --log-driver=syslog --log-opt syslog-address=udp://<IP ADDRESS WHERE syslog-to-metron IS RUNNING>:10514 --log-opt syslog-format=rfc3164 -d redis
```

### Check you application logs

Check you application logs. The service logs must appear mixed with the application logs.

```
cf logs <APP-NAME>
```

## Contributing

In the spirit of [free software](http://www.fsf.org/licensing/essays/free-sw.html), **everyone** is encouraged to help improve this project.

Here are some ways *you* can contribute:

* by using alpha, beta, and prerelease versions
* by reporting bugs
* by suggesting new features
* by writing or editing documentation
* by writing specifications
* by writing code (**no patch is too small**: fix typos, add comments, clean up inconsistent whitespace)
* by refactoring code
* by closing [issues](https://github.com/frodenas/syslog-to-metron/issues)
* by reviewing patches

### Submitting an Issue

We use the [GitHub issue tracker](https://github.com/frodenas/syslog-to-metron/issues) to track bugs and features. Before submitting a bug report or feature request, check to make sure it hasn't already been submitted. You can indicate support for an existing issue by voting it up. When submitting a bug report, please include a [Gist](http://gist.github.com/) that includes a stack trace and any details that may be necessary to reproduce the bug. Ideally, a bug report should include a pull request with failing specs.

### Submitting a Pull Request

1. Fork the project.
2. Create a topic branch.
3. Implement your feature or bug fix.
4. Commit and push your changes.
5. Submit a pull request.

## Copyright

Copyright (c) 2016 Ferran Rodenas. See [LICENSE](https://github.com/frodenas/syslog-to-metron/blob/master/LICENSE) for details.
