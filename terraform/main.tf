provider "aws" {
  region  = local.region
  profile = var.profile
}

data "aws_vpc" "default" {
  default = true
}

data "aws_subnets" "default" {
  filter {
    name   = "vpc-id"
    values = [data.aws_vpc.default.id]
  }
}

data "aws_iam_role" "lab_role" {
  name = "LabRole"
}

resource "aws_iam_instance_profile" "lab_instance_profile" {
  name = "lab-instance-profile"
  role = data.aws_iam_role.lab_role.name
}

resource "aws_key_pair" "node-key" {
  key_name   = "cluster-keypair"
  public_key = file(local.key_file_name)
}

resource "aws_instance" "master" {
  ami                    = local.ami
  instance_type          = local.instance_type
  key_name               = aws_key_pair.node-key.key_name
  vpc_security_group_ids = [aws_security_group.example.id]
  user_data_base64 = base64encode("${templatefile("scripts/initMaster.sh", {
    KUBERNETES_VERSION = "1.29.0-1.1",
    CRIO_OS            = "xUbuntu_22.04",
    CRIO_VERSION       = "1.27", #volver a poner 1.28 cuando se descaiga
    NODE_NAME          = "master",
    POD_NETWORK_CIDR   = "192.168.0.0/16"
  })}")

  iam_instance_profile = aws_iam_instance_profile.lab_instance_profile.name

  root_block_device {
    volume_size = 30
    volume_type = "gp3"
  }

  tags = {
    Name = "master-node"
  }
}

resource "aws_instance" "node" {
  count                  = 1
  ami                    = local.ami
  instance_type          = local.instance_type
  key_name               = aws_key_pair.node-key.key_name
  vpc_security_group_ids = [aws_security_group.example.id]
  user_data_base64 = base64encode("${templatefile("scripts/initWorker.sh", {
    HOSTNAME           = "worker-node-${count.index + 1}",
    CRIO_OS            = "xUbuntu_22.04",
    CRIO_VERSION       = "1.27", #volver a poner 1.28 cuando se descaiga
    KUBERNETES_VERSION = "1.29.0-1.1",
  })}")

  root_block_device {
    volume_size = 30
    volume_type = "gp3"
  }


  tags = {
    Name = "worker-node-${count.index + 1}"
  }
}

resource "aws_security_group" "example" {
  name        = "example-security-group"
  description = "Allow all inbound and outbound traffic"

  ingress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

# The map here can come from other supported configurations
# like locals, resource attribute, map() built-in, etc.
variable "my_secretss" {
  default = {
    key1 = "value1"
    key2 = "value2"
  }
  sensitive     = true
  type = map(string)
}

resource "aws_secretsmanager_secret" "my_secretss" {
  name = "my_secretss"
}

resource "aws_secretsmanager_secret_version" "my_secretss" {
  secret_id     = aws_secretsmanager_secret.my_secretss.id
  secret_string = jsonencode(var.my_secretss)
  
}

resource "aws_secretsmanager_secret_policy" "secret_policy" {
  secret_arn = aws_secretsmanager_secret.my_secretss.arn
  policy     = jsonencode({
        Version = "2012-10-17",
        Statement = [
            {
                Effect = "Allow",
                Principal = {
                    AWS = data.aws_iam_role.lab_role.arn
                },
                Action = [
                    "secretsmanager:GetSecretValue",
                    "secretsmanager:DescribeSecret"
                ],
                Resource = aws_secretsmanager_secret.my_secretss.arn
            }
        ]
    })
}

resource "aws_vpc_endpoint" "secrets_manager" {
  vpc_id            = data.aws_vpc.default.id
  service_name      = "com.amazonaws.${local.region}.secretsmanager"
  vpc_endpoint_type   = "Interface"

  private_dns_enabled = true
  subnet_ids         = data.aws_subnets.default.ids

  security_group_ids = [aws_security_group.example.id]

  dns_options {
    dns_record_ip_type = "ipv4"
    private_dns_only_for_inbound_resolver_endpoint = true
  }

  policy = jsonencode({
      Version = "2012-10-17",
      Statement = [
        {
          Effect = "Allow"
          Action = ["secretsmanager:GetSecretValue", "secretsmanager:DescribeSecret"]
          Resource = aws_secretsmanager_secret.my_secretss.arn
          Principal = {
              AWS = data.aws_iam_role.lab_role.arn
          }
        }
      ]
  })
}

output "ssh_commands_to_worker_nodes" {
  value       = [for public_dns in aws_instance.node.*.public_dns : "ssh ubuntu@${public_dns}"]
  description = "SSH commands to connect to Kubernetes worker nodes"
}

output "ssh_commands_to_master_node" {
  value       = "ssh ubuntu@${aws_instance.master.public_dns}"
  description = "SSH commands to connect to Kubernetes master node"
}
