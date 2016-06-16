A ridiculously awesome LRSC to MQTT bridge written in GO. 

# Note!

This repository is no longer maintained. We are releasing this code in the hope that it will be useful to the community. 
The code was developed against an older version of LRSC that used a basic socket for communication. Since we developed it, the LRSC interface has changed to use websockets, so the code will 
need updating to work against the newer version. 

# What the bridge does 

This bridge integrates the [Long-Range Signaling and Control (LRSC)](http://www.research.ibm.com/labs/zurich/ics/lrsc/) platform into Bluemix. Based on the LoRa technology, LRSC allows very low power devices to communicate wirelessly over a range of several miles.

Sensors connect to a physical gateway device, which bridges between the LoRa wireless technology and standard IP networking. The gateway then connects to a centralised LRSC server, which manages the network. The server connects to our bridge, which then passes messages to the IBM Internet of Things Foundation (IoTF) service on Bluemix.

![lrsc overview diagram](http://garage.mybluemix.net/posts/lrsc/overview.png)

The [Bluemix Garage blog](http://garage.mybluemix.net/posts/lrsc/) has lots more information. 

# Deploying to Bluemix

You will need the to have the latest stable [cf cli](https://github.com/cloudfoundry/cli#downloads) installed on your host and available in your terminal. [See Bluemix documentation for more details](https://www.ng.bluemix.net/docs/#starters/install_cli.html).

You will need to create an **Internet of Things** service called **iotf** in the Bluemix org and space that you will be deploying this application to. The **iotf** name is important, the app's manifest requires this. For the time being, **IoT** is only available in Bluemix **US South**, ensure you are in the correct region when creating this service.

You will need to obtain a key archive from the [LRSC Application Router web interface](https://lrsc.ch/ssc/help.html#ssc-getting-started).  After logging in, click the **Setup** link at the top and then **Download key archive**.

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


Build the bridge and then place **client.cert** and **client.key** into the extracted **bluemixgarage_lrsc-bridge-master** folder, alongside **manifest.yml**.

You will need to rename the **host** entry in the **manifest.yml** file as it will clash with our own **lrsc-bridge** instance. Host names must be unique per Bluemix region. You might want to use the same name for the **name** entry, but it's not a requirement.

From your terminal, navigate to this folder, `cf login` with your Bluemix credentials and then `cf push`.

To verify that the bridge is working correctly:

1. Visit the Bluemix URL where the app is deployed (shown after `cf push`) to view the status page
1. From the Bluemix Dashboard, open the **iotf** service, then click the **LAUNCH** button. You should see devices appear in the list.
