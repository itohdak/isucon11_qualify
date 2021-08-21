sudo rm /var/log/nginx/access.log
sudo rm /var/log/mysql/mariadb-slow.log
sudo systemctl restart nginx mysql mysqld isucondition.go
