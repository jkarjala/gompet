mkdir -p tmp
mkdir -p tmp/windows
mkdir -p tmp/linux
mkdir -p tmp/darwin

for tool in $(find . -type d -name "gompet-*") 
do
  echo Building $tool
  GOOS=windows GOARCH=amd64 go build -o tmp/windows/${tool}.exe $tool
  GOOS=linux GOARCH=amd64 go build -o tmp/linux/$tool $tool
  GOOS=darwin GOARCH=amd64 go build -o tmp/darwin/$tool $tool
done
cd tmp
zip -j gompet-windows-amd64.zip windows/*
zip -j gompet-linux-amd64.zip linux/*
zip -j gompet-darwin-amd64.zip darwin/*
