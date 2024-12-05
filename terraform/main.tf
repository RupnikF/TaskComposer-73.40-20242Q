provider "aws" {
  region  = local.region
  profile = var.profile
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
    CRIO_VERSION       = "1.28",
    NODE_NAME          = "master",
    POD_NETWORK_CIDR   = "192.168.0.0/16"
  })}")

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
    CRIO_VERSION       = "1.28",
    KUBERNETES_VERSION = "1.29.0-1.1",
  })}")


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

output "ssh_commands_to_worker_nodes" {
  value       = [for public_dns in aws_instance.node.*.public_dns : "ssh ubuntu@${public_dns}"]
  description = "SSH commands to connect to Kubernetes worker nodes"
}

output "ssh_commands_to_master_node" {
  value       = "ssh ubuntu@${aws_instance.master.public_dns}"
  description = "SSH commands to connect to Kubernetes master node"
}
