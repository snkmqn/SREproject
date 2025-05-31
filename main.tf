provider "aws" {
  region = var.region
}

resource "aws_instance" "app" {
  ami           = var.ami_id
  instance_type = "t3.micro"
  key_name      = var.key_name

  tags = {
    Name = "iac-app"
  }

  provisioner "local-exec" {
    command = "echo ${self.public_ip} > ../ansible/inventory.ini"
  }

  provisioner "file" {
    source      = "./bin/productservice"
    destination = "/home/ubuntu/productservice"
  }

  provisioner "remote-exec" {
      inline = [
        "sudo apt update",
        "sudo apt install -y mongodb redis-server docker.io",
        "sudo systemctl start mongodb",
        "sudo systemctl enable mongodb",
        "sudo systemctl start redis-server",
        "sudo systemctl enable redis-server",
        "chmod +x /home/ubuntu/productservice",
        "nohup /home/ubuntu/productservice > /home/ubuntu/productservice.log 2>&1 &"
        "sudo docker run -d -p 80:8080 yourdockerhubuser/productservice"

      ]
  }

  connection {
    type        = "ssh"
    user        = "ubuntu"
    private_key = file(var.private_key_path)
    host        = self.public_ip
  }
}

output "public_ip" {
  value = aws_instance.app.public_ip
}