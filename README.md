A ridiculously awesome LRSC to MQTT bridge written in GO.

# Deploying to Bluemix

You will need the to have the latest stable [cf cli](https://github.com/cloudfoundry/cli#downloads) installed on your host and available in your terminal. [See Bluemix documentation for more details](https://www.ng.bluemix.net/docs/#starters/install_cli.html).

You will need to create an **Internet of Things** service called **iotf** in the Bluemix org and space that you will be deploying this application to. The **iotf** name is important, the app's manifest requires this. For the time being, **IoT** is only available in Bluemix **US South**, ensure you are in the correct region when creating this service.

You will need to obtain a key archive from the [LRSC Application Router web interface](https://dev.lrsc.ch/).  After logging in, click the **Setup** link at the top and then **Download key archive**.

After expanding the archive, you should have the following files.  These are the ones you need:
```
.
├── certs
│   ├── AA-AA-AA-AA-FF-FF-FF-FF.CLIENT.cert                <-- rename to client.cert
│   ├── AA-AA-AA-AA-FF-FF-FF-FF.CLIENT.cert.trust.jks
│   ├── AA-AA-AA-AA-FF-FF-FF-FF.cert
│   ├── CA.cert
│   └── CA.cert.der
└── private
    ├── AA-AA-AA-AA-FF-FF-FF-FF.CLIENT.key                 <-- rename to client.key
    └── AA-AA-AA-AA-FF-FF-FF-FF.CLIENT.key.jks
```

In a separate folder, place these 4 files:

1. **lrsc-bridge** - from this repository, this is the latest stable go binary compiled for Linux 64bit (the Bluemix default)
1. **manifest.yml** - from this repository, used by the cf cli when pushing to Bluemix
1. **client.cert** - download from LRSC web, this is your LRSC router TLS certificate
1. **client.key** - download from LRSC web, this is your LRSC router TLS private key

You will need to rename the **host** entry in the **manifest.yml** file as it will clash with our own **lrsc-bridge** instance. Host names must be unique per Bluemix region. You might want to use the same name for the **name** entry, but it's not a requirement.

Once you have the above files in place, from your terminal, navigate to this folder, `cf login` with your Bluemix credentials and then `cf push`.
