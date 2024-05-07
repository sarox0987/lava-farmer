# Install git

```
sudo apt-get install git
```

# Install go

```
VERSION="1.21.6"
ARCH="amd64"
curl -O -L "https://golang.org/dl/go${VERSION}.linux-${ARCH}.tar.gz"
tar -xf "go${VERSION}.linux-${ARCH}.tar.gz"
sudo chown -R root:root ./go
sudo mv -v go /usr/local
export GOPATH=$HOME/go
export PATH=$PATH:/usr/local/go/bin:$GOPATH/bin
source ~/.bash_profile
go version
```

# Clone Repository

```
git clone https://github.com/sarox0987/lava-farmer.git
cd lava-farmer
```

# Set RPCs
```
nano env.json
```

# Install and create a screen session
```
sudo apt install screen
screen -S lava
```

# Run the script
``` 
go run main.go
```

Exit from the screen with ```Ctrl + A + D```
