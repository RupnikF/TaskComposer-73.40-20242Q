import json
from typing import List, Optional, Tuple
import boto3
import os, sys, threading
from confluent_kafka import Consumer, KafkaError, Producer

from opentelemetry.sdk.resources import Resource
from flask import Flask, jsonify
import logging
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.export import (
    BatchSpanProcessor,
    ConsoleSpanExporter,
)
from opentelemetry import trace
from opentelemetry.trace.propagation.tracecontext import TraceContextTextMapPropagator
from opentelemetry.exporter.otlp.proto.http.trace_exporter import OTLPSpanExporter

provider = TracerProvider(resource=Resource.create({"service.name": "s3-service"}))
processor = BatchSpanProcessor(ConsoleSpanExporter())
processor_prod = BatchSpanProcessor(OTLPSpanExporter())
provider.add_span_processor(processor)
provider.add_span_processor(processor_prod)

tracer = trace.get_tracer("s3-service")

logger = logging.getLogger("s3-service")

os.environ["PYTHONUNBUFFERED"] = "1"

log_file_path = os.getenv("LOG_FILE_PATH", "./s3-service.log")
handler = logging.FileHandler(log_file_path, mode="a")
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
        "aws_access_key": "access-key",
        "aws_secret_key": "secret-key",
        "aws_region": "region"
    }
    "traceId": "",
}
"""

# Kafka Configuration
KAFKA_BROKER = f"{os.getenv('KAFKA_HOST')}:{os.getenv('KAFKA_PORT')}"
INPUT_TOPIC = os.getenv("INPUT_TOPIC")
OUTPUT_TOPIC = os.getenv("OUTPUT_TOPIC")
SERVICE_NAME = os.getenv("SERVICE_NAME")
FILES_BASE_PATH = os.getenv("FILES_BASE_PATH", "/data")

producer_config = {
    "bootstrap.servers": KAFKA_BROKER
}
producer = Producer(producer_config)

class S3Service():

    def __init__(self, producer=None, consumer=None, input_topic=INPUT_TOPIC, output_topic=OUTPUT_TOPIC):
        if producer is None:
            producer_config = {
                "bootstrap.servers": KAFKA_BROKER
            }
            self.producer = Producer(producer_config)
        else:    
            self.producer = producer
        if consumer is None:
            consumer_config = {
                "bootstrap.servers": KAFKA_BROKER,
                "group.id": "s3service",
                "auto.offset.reset": "earliest"
            }
            self.consumer = Consumer(consumer_config)
        else:
            self.consumer = consumer
        self.input_topic = input_topic
        self.output_topic = output_topic
        self.consumer.subscribe([self.input_topic])

    def create_s3_client(self, access_key, secret_key, session_token, region):
        """Creates an S3 client using the provided credentials."""
        return boto3.Session(
            aws_access_key_id=access_key,
            aws_secret_access_key=secret_key,
            region_name=region,
            aws_session_token=session_token,
        ).client('s3')

    def write_to_kafka(self, execution_id, status, path=None, error=None):
        """Writes a result message to Kafka."""
        topic = self.output_topic
        body = {
            "status": status
        }
        if path:
            body["path"] = path
        if error:
            body["error_msg"] = error
        message = json.dumps({
            "executionId": execution_id,
            "outputs": body
        }).encode("utf-8")
        try:
            logger.info(f"Writing to Kafka: {message}")
            self.producer.produce(topic=topic, message=message)
            self.producer.flush()
        except Exception as e:
            logger.info(f"Error writing to Kafka: {e}")

    def download_from_s3(self, execution_id, s3_client, bucket_name, s3_key, file_path):
        """Downloads a file from S3 to the local filesystem."""
        try:
            logger.info(f"Downloading from S3: bucket={bucket_name}, key={s3_key} to {file_path}")
            s3_client.download_file(bucket_name, s3_key, file_path)
            logger.info("Download successful!")
            self.write_to_kafka(execution_id, "success", path=file_path)
        except Exception as e:
            error_msg = f"Error downloading from S3: {e}"
            logger.error(error_msg)
            self.write_to_kafka(execution_id, "error", error=error_msg)

    def upload_to_s3(self, execution_id, s3_client, bucket_name, s3_key, file_path):
        """Uploads a local file to S3."""
        try:
            if not os.path.isfile(file_path):
                logger.error(f"File does not exist: {file_path}")
                self.write_to_kafka(execution_id, "error", error=f"File does not exist: {file_path}")
                return
            logger.info(f"Uploading to S3: {file_path} to bucket={bucket_name}, key={s3_key}")
            s3_client.upload_file(file_path, bucket_name, s3_key)
            logger.info("Upload successful!")
            self.write_to_kafka(execution_id, "success", path=f"s3://{bucket_name}/{s3_key}")
        except Exception as e:
            error_msg = f"Error uploading to S3: {e}"
            logger.error(error_msg)
            self.write_to_kafka(execution_id, "error", error=error_msg)

    def extract_ctx(self, kafka_message):
        headers: Optional[List[Tuple[str, bytes]]] = kafka_message.headers()
        if headers is None:
            return None
        headers_map = {k: v.decode('utf-8') for k, v in headers}
        return TraceContextTextMapPropagator().extract(headers_map)

    # TODO: test
    def inject_ctx(self, span, kafka_message):
        headers = {}
        TraceContextTextMapPropagator().inject(span.context, headers)
        kafka_message.set_headers([(k, v.encode('utf-8')) for k, v in headers.items()])

    def process_message(self, message):
        ctx = self.extract_ctx(message)
        with tracer.start_as_current_span("process_message", context=ctx) as span:
            try:
                data = json.loads(message.value().decode("utf-8"))
                span.set_attribute("data", data)
                execution_id = data.get("executionId")
                task = data.get("taskName")
                inputs = data.get("inputs", None)
                if inputs:
                    file_path = f"{FILES_BASE_PATH}/{inputs.get("file_path")}"
                    bucket_name = inputs.get("bucket_name")
                    s3_key = inputs.get("s3_key")
                    aws_access_key = inputs.get("aws_access_key")
                    aws_secret_key = inputs.get("aws_secret_key")
                    aws_region = inputs.get("aws_region")
                    aws_session_token = inputs.get("aws_session_token")

                # Validate required fields
                if inputs is None or not all([execution_id, task, file_path, bucket_name, s3_key, aws_access_key, aws_secret_key, aws_region]):
                    exc_msg = "Invalid message: missing required fields"
                    span.record_exception(Exception(exc_msg))
                    logger.error(exc_msg)
                    self.write_to_kafka(execution_id, "error", error=exc_msg)
                    return
                
                if task.lower() not in ["download", "upload"]:
                    exc_msg = f"Unsupported task: {task}"
                    span.record_exception(Exception(exc_msg))
                    logger.error(exc_msg)
                    self.write_to_kafka(execution_id, "error", error=exc_msg)
                    return

                # Create S3 client
                s3_client = self.create_s3_client(aws_access_key, aws_secret_key, aws_session_token, aws_region)

                # Perform operation
                if task.lower() == "download":
                    self.download_from_s3(execution_id, s3_client, bucket_name, s3_key, file_path)
                elif task.lower() == "upload":
                    self.upload_to_s3(execution_id, s3_client, bucket_name, s3_key, file_path)
                    
                # inject_ctx(span, message) when sending the message

            except json.JSONDecodeError:
                exc_msg = "Invalid message format: Not a valid JSON"
                logger.error(exc_msg)
                if execution_id:
                    self.write_to_kafka(execution_id, "error", error="Invalid message format: Not a valid JSON")
            except Exception as e:
                exc_msg = f"Error processing message: {e}"
                logger.error(exc_msg)
                if execution_id:
                    self.write_to_kafka(execution_id, "error", error=exc_msg)


    def consume_kafka_messages(self):
        logger.info("Listening for messages...")
        try:
            while True:
                # TODO: Single threaded
                msg = self.consumer.poll(1.0)

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
                self.process_message(msg)

        except KeyboardInterrupt:
            logger.info("Stopping Kafka consumer...")
        finally:
            self.consumer.close()

@app.route('/healthz')
def healthz():
    return jsonify({"status": "ok"}), 200

if __name__ == "__main__":
    service = S3Service()
    kafka_thread = threading.Thread(target=S3Service.consume_kafka_messages, daemon=True)
    kafka_thread.start()
    app.run(host='0.0.0.0', port=80)
