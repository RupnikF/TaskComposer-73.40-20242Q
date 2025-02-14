import unittest
from unittest.mock import patch, MagicMock
from service import S3Service

INPUT_TOPIC = 'input_topic'
OUTPUT_TOPIC = 'output_topic'

# se corre con `python -m unittest test_s3_service.py`
class TestS3Service(unittest.TestCase):

    @patch('service.Producer')
    @patch('service.Consumer')
    @patch('service.logger')
    def setUp(self, mock_logger, mock_consumer, mock_producer):
        self.mock_producer = mock_producer
        self.mock_consumer = mock_consumer
        self.mock_logger = mock_logger
        self.service = S3Service(producer=self.mock_producer, consumer=self.mock_consumer, input_topic=INPUT_TOPIC, output_topic=OUTPUT_TOPIC)

    @patch('service.boto3.Session')
    def test_create_s3_client(self, mock_boto_session):
        mock_session_instance = mock_boto_session.return_value
        mock_client = mock_session_instance.client.return_value

        s3_client = self.service.create_s3_client('access_key', 'secret_key', 'session_token', 'region')

        mock_boto_session.assert_called_once_with(
            aws_access_key_id='access_key',
            aws_secret_access_key='secret_key',
            region_name='region',
            aws_session_token='session_token'
        )
        mock_session_instance.client.assert_called_once_with('s3')
        self.assertEqual(s3_client, mock_client)

    @patch('service.json.dumps')
    def test_write_to_kafka(self, mock_json_dumps):
        self.service.write_to_kafka('execution_id', 'status', 'path', 'error')

        mock_json_dumps.assert_called_once()
        self.mock_producer.produce.assert_called_once()
        self.mock_producer.produce.assert_called_once_with(
            topic=OUTPUT_TOPIC,
            message=mock_json_dumps.return_value.encode('utf-8')
        )
        self.mock_producer.flush.assert_called_once()
    
    @patch('service.boto3.Session')
    def test_download_from_s3(self, mock_boto_session):
        mock_client = mock_boto_session.return_value.client.return_value

        with patch('builtins.open', unittest.mock.mock_open()) as mock_file, \
             patch.object(self.service, 'write_to_kafka') as mock_write_to_kafka:
            self.service.download_from_s3('execution_id', mock_client, 'bucket_name', 's3_key', 'file_path')

        mock_client.download_file.assert_called_once_with('bucket_name', 's3_key', 'file_path')
        mock_write_to_kafka.assert_called_with('execution_id', 'success', path='file_path')

    
    @patch('service.boto3.Session')
    def test_upload_to_s3(self, mock_boto_session):
        mock_client = mock_boto_session.return_value.client.return_value
        with patch('os.path.isfile', return_value=True), \
             patch('builtins.open', unittest.mock.mock_open()) as mock_file, \
             patch.object(self.service, 'write_to_kafka') as mock_write_to_kafka:
            self.service.upload_to_s3('execution_id', mock_client, 'bucket_name', 's3_key', 'file_path')

        mock_client.upload_file.assert_called_once_with('file_path', 'bucket_name', 's3_key')
        mock_write_to_kafka.assert_called_with('execution_id', 'success', path='s3://bucket_name/s3_key')

    @patch('service.boto3.Session')
    def test_upload_to_s3_file_not_exist(self, mock_boto_session):
        mock_client = mock_boto_session.return_value.client.return_value
        with patch('os.path.isfile', return_value=False):
            self.service.upload_to_s3('execution_id', mock_client, 'bucket_name', 's3_key', 'file_path')

        mock_client.upload_file.assert_not_called()

    
    def test_process_message(self):
        mock_message = MagicMock()
        mock_message.value.return_value = b'{"executionId": "123", "taskName": "download", "inputs": {"file_path": "path/to/file", "bucket_name": "bucket-name", "s3_key": "path/in/bucket", "aws_access_key": "access-key", "aws_secret_key": "secret-key", "aws_region": "region"}}'

        with patch.object(self.service, 'download_from_s3') as mock_download, \
             patch.object(self.service, 'upload_to_s3') as mock_upload:
            self.service.process_message(mock_message)

        mock_download.assert_called_once()
        mock_upload.assert_not_called()
    
    def test_process_message_missing_fields(self):
        mock_message = MagicMock()
        mock_message.value.return_value = b'{"executionId": "123", "taskName": "download"}'

        with patch.object(self.service, 'download_from_s3') as mock_download, \
             patch.object(self.service, 'upload_to_s3') as mock_upload, \
             patch.object(self.service, 'write_to_kafka') as mock_write_to_kafka:
            self.service.process_message(mock_message)

        mock_download.assert_not_called()
        mock_upload.assert_not_called()
        mock_write_to_kafka.assert_called_with('123', 'error', error='Invalid message: missing required fields')

    def test_process_message_invalid_task(self):
        mock_message = MagicMock()
        mock_message.value.return_value = b'{"executionId": "123", "taskName": "invalid_task", "inputs": {"file_path": "path/to/file", "bucket_name": "bucket-name", "s3_key": "path/in/bucket", "aws_access_key": "access-key", "aws_secret_key": "secret-key", "aws_region": "region"}}'

        with patch.object(self.service, 'download_from_s3') as mock_download, \
             patch.object(self.service, 'upload_to_s3') as mock_upload, \
             patch.object(self.service, 'write_to_kafka') as mock_write_to_kafka:
            self.service.process_message(mock_message)

        mock_download.assert_not_called()
        mock_upload.assert_not_called()
        mock_write_to_kafka.assert_called_with('123', 'error', error='Unsupported task: invalid_task')
    
if __name__ == '__main__':
    unittest.main()