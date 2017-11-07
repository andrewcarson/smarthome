import boto3
import logging
import json
import time
import uuid

# Setup logger
logger = logging.getLogger()
logger.setLevel(logging.INFO)

def lambda_handler(request, context):
    send_message()

    try:
        logger.info("Directive:")
        logger.info(json.dumps(request, indent=4, sort_keys=True))

        version = get_directive_version(request)

        if version == "3":
            logger.info("Received v3 directive!")
            if request["directive"]["header"]["name"] == "Discover":
                response = handle_discovery_v3(request)
            else:
                response = handle_non_discovery_v3(request)

        else:
            raise Exception("Received v2 directive!")

        logger.info("Response:")
        logger.info(json.dumps(response, indent=4, sort_keys=True))

        return response
    except ValueError as error:
        logger.error(error)
        raise


# utility functions
def get_utc_timestamp(seconds=None):
    return time.strftime("%Y-%m-%dT%H:%M:%S.00Z", time.gmtime(seconds))

def get_uuid():
    return str(uuid.uuid4())

# v3 handlers
def handle_discovery_v3(request):
    endpoints = [
        {
            "endpointId": "endpoint-samsung-tv",
            "manufacturerName": "Samsung",
            "friendlyName": "TV",
            "description": "Samsung TV",
            "displayCategories": [ 'OTHER' ],
            "cookie": {
            },
            "capabilities": [
                {
                    "type": "AlexaInterface",
                    "interface": "Alexa.InputController",
                    "version": "3",
                    "properties": {
                        "supported": [
                            {
                                "name": "input"
                            }
                        ],
                        "proactivelyReported": False,
                        "retrievable": False
                    }
                },
                {
                    "type": "AlexaInterface",
                    "interface": "Alexa.EndpointHealth",
                    "version": "3",
                    "properties": {
                        "supported":[
                            { "name":"connectivity" }
                        ],
                        "proactivelyReported": False,
                        "retrievable": True
                    }
                },
                {
                    "type": "AlexaInterface",
                    "interface": "Alexa",
                    "version": "3"
                }
            ]
        }
    ]

    response = {
        "event": {
            "header": {
                "namespace": "Alexa.Discovery",
                "name": "Discover.Response",
                "payloadVersion": "3",
                "messageId": get_uuid()
            },
            "payload": {
                "endpoints": endpoints
            }
        }
    }
    return response

def send_message():
    sqs = boto3.resource('sqs')
    queue = sqs.get_queue_by_name(QueueName='alexa-smarthome')
    response = queue.send_message(MessageBody="from the lambda")

def handle_non_discovery_v3(request):
    request_namespace = request["directive"]["header"]["namespace"]
    request_name = request["directive"]["header"]["name"]

    if request_namespace == "Alexa.InputController":
        if request_name == "SelectInput":
            response = {
                "context": {
                    "properties": [
                        {
                            "namespace": "Alexa.InputController",
                            "name": "input",
                            "value": 'HDMI 1',
                            "timeOfSample": get_utc_timestamp(),
                            "uncertaintyInMilliseconds": 500
                        }
                    ]
                },
                "event": {
                    "header": {
                        "namespace": "Alexa",
                        "name": "Response",
                        "payloadVersion": "3",
                        "messageId": get_uuid(),
                        "correlationToken": request["directive"]["header"]["correlationToken"]
                    },
                    "endpoint": {
                        "endpointId": request["directive"]["endpoint"]["endpointId"]
                    },
                    "payload": {}
                }
            }

            return response

    elif request_namespace == "Alexa.Authorization":
        if request_name == "AcceptGrant":
            response = {
                "event": {
                    "header": {
                        "namespace": "Alexa.Authorization",
                        "name": "AcceptGrant.Response",
                        "payloadVersion": "3",
                        "messageId": "5f8a426e-01e4-4cc9-8b79-65f8bd0fd8a4"
                    },
                    "payload": {}
                }
            }
            return response

    # other handlers omitted in this example

# v3 utility functions
def get_directive_version(request):
    try:
        return request["directive"]["header"]["payloadVersion"]
    except:
        try:
            return request["header"]["payloadVersion"]
        except:
            return "-1"
