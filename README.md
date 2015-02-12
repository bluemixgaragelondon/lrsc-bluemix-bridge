A ridiculously awesome LRSC to MQTT bridge written in GO.

# Deploying to Bluemix

You will need the to have the latest stable [cf cli](https://github.com/cloudfoundry/cli#downloads) installed on your host and available in your terminal. [See Bluemix documentation for more details](https://www.ng.bluemix.net/docs/#starters/install_cli.html).

You will need to create an **Internet of Things** service called **iotf** in the Bluemix org and space that you will be deploying this application to. The **iotf** name is important, the app's manifest requires this. For the time being, **IoT** is only available in Bluemix **US South**, ensure you are in the correct region when creating this service.

You will need to obtain a key archive from the [LRSC Application Router web interface](https://dev.lrsc.ch/).  After logging in, click the **Setup** link at the top and then **Download key archive**.

After expanding the archive, you should have the following files:
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

You will want to download the **lrsc-bridge** app from [Jazz Hub](https://hub.jazz.net/project/bluemixgarage/lrsc-bridge/overview) as a zip archive. There will be a download button next to the **Branch: master** dropdown.

Extract the zip file and then place **client.cert** and **client.key** into the extracted **bluemixgarage_lrsc-bridge-master** folder, alongside **manifest.yml**.

You will need to rename the **host** entry in the **manifest.yml** file as it will clash with our own **lrsc-bridge** instance. Host names must be unique per Bluemix region. You might want to use the same name for the **name** entry, but it's not a requirement.

From your terminal, navigate to this folder, `cf login` with your Bluemix credentials and then `cf push`.

To verify that the bridge is working correctly:

1. Visit the Bluemix URL where the app is deployed (shown after `cf push`) to view the status page
1. From the Bluemix Dashboard, open the **iotf** service, then click the **LAUNCH** button. You should see devices appear in the list.
