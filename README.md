This is a Go program that looks in a directory for picture files. 
It randomly selects one of them, and then publishes info about it 
via a built in web server. 
This is used as a photo provider for a picture viewer setup. 

Install:
sudo cp picFetch386 /usr/local/bin/
sudo cp picFetch.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl start picFetch
sudo systemctl status picFetch

Config functions is inspired by https://dev.to/koddr/let-s-write-config-for-your-golang-web-app-on-right-way-yaml-5ggp.


Build: GOOS=linux GOARCH=386 go build -o picFetch