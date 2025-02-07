import json
import boto3
import os, sys, threading
from confluent_kafka import Consumer, KafkaError
from flask import Flask, jsonify
import logging
logger = logging.getLogger("s3-service")

os.environ["PYTHONUNBUFFERED"] = "1"

handler = logging.FileHandler("/var/log/s3-service.log", mode="a")
handler2 = logging.StreamHandler(sys.stdout)
handler.setLevel(logging.INFO)
handler2.setLevel(logging.INFO)

formatter = logging.Formatter("%(asctime)s - %(name)s - %(levelname)s - %(message)s")
handler.setFormatter(formatter)
handler2.setFormatter(formatter)

logger.setLevel(logging.INFO)  # Ensure logger level is set
logger.addHandler(handler)
logger.addHandler(handler2)

app = Flask(__name__)

"""
{
    "executionId": 123,
    "taskName": "download|upload",
    "inputs": {
        "file_path": "path/to/file",
        "bucket_name": "bucket-name",
        "s3_key": "path/in/bucket",
        "aws_access_key": "access-key
        "aws_secret_key": "secret-key
        "aws_region": "region"
    }
    "traceId": "123",
}
"""

# Kafka Configuration
KAFKA_BROKER = f"{os.getenv('KAFKA_HOST')}:{os.getenv('KAFKA_PORT')}"
INPUT_TOPIC = os.getenv("INPUT_TOPIC")
OUTPUT_TOPIC = os.getenv("OUTPUT_TOPIC")
SERVICE_NAME = os.getenv("SERVICE_NAME")
FILES_BASE_PATH = os.getenv("FILES_BASE_PATH", "/data")

def create_s3_client(access_key, secret_key, session_token, region):
    """Creates an S3 client using the provided credentials."""
    return boto3.Session(
        aws_access_key_id=access_key,
        aws_secret_access_key=secret_key,
        region_name=region,
        aws_session_token=session_token,
    ).client('s3')

def download_from_s3(s3_client, bucket_name, s3_key, file_path):
    """Downloads a file from S3 to the local filesystem."""
    try:
        logger.info(f"Downloading from S3: bucket={bucket_name}, key={s3_key} to {file_path}")
        s3_client.download_file(bucket_name, s3_key, file_path)
        logger.info("Download successful!")
    except Exception as e:
        logger.info(f"Error downloading from S3: {e}")

def upload_to_s3(s3_client, bucket_name, s3_key, file_path):
    """Uploads a local file to S3."""
    try:
        if not os.path.isfile(file_path):
            logger.info(f"File does not exist: {file_path}")
            return
        logger.info(f"Uploading to S3: {file_path} to bucket={bucket_name}, key={s3_key}")
        s3_client.upload_file(file_path, bucket_name, s3_key)
        logger.info("Upload successful!")
    except Exception as e:
        logger.info(f"Error uploading to S3: {e}")

def process_message(message):
    try:
        data = json.loads(message)
        task = data.get("taskName")
        inputs = data.get("inputs")
        file_path = f"{FILES_BASE_PATH}/{inputs.get("file_path")}"
        bucket_name = inputs.get("bucket_name")
        s3_key = inputs.get("s3_key")
        aws_access_key = inputs.get("aws_access_key")
        aws_secret_key = inputs.get("aws_secret_key")
        aws_region = inputs.get("aws_region")
        aws_session_token = inputs.get("aws_session_token")

        # Validate required fields
        if not all([task, file_path, bucket_name, s3_key, aws_access_key, aws_secret_key, aws_region]):
            logger.info("Invalid message: missing required fields")
            return
        
        if task.lower() not in ["download", "upload"]:
            logger.info(f"Unsupported task: {task}")
            return

        # Create S3 client
        s3_client = create_s3_client(aws_access_key, aws_secret_key, aws_session_token, aws_region)

        # Perform operation
        if task.lower() == "download":
            download_from_s3(s3_client, bucket_name, s3_key, file_path)
        elif task.lower() == "upload":
            upload_to_s3(s3_client, bucket_name, s3_key, file_path)

    except json.JSONDecodeError:
        logger.info("Invalid message format: Not a valid JSON")
    except Exception as e:
        logger.info(f"Error processing message: {e}")

def consume_kafka_messages():
    consumer_config = {
        "bootstrap.servers": KAFKA_BROKER,
        "group.id": "s3service",
        "auto.offset.reset": "earliest"
    }

    consumer = Consumer(consumer_config)
    consumer.subscribe([INPUT_TOPIC])

    logger.info("Listening for messages...")
    try:
        while True:
            msg = consumer.poll(1.0)

            if msg is None:
                continue
            if msg.error():
                if msg.error().code() == KafkaError._PARTITION_EOF:
                    continue
                else:
                    logger.info(f"Error: {msg.error()}")
                    continue

            # Process the message
            logger.info(f"Received message: {msg.value().decode('utf-8')}")
            process_message(msg.value().decode("utf-8"))

    except KeyboardInterrupt:
        logger.info("Stopping Kafka consumer...")
    finally:
        consumer.close()

@app.route('/healthz')
def healthz():
    return jsonify({"status": "ok"}), 200

if __name__ == "__main__":
    logger.info(f"HELLO")
    kafka_thread = threading.Thread(target=consume_kafka_messages, daemon=True)
    kafka_thread.start()
    app.run(host='0.0.0.0', port=80)
