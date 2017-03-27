echo "正在编译win64.."
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build  -o release/builds/win64/biu.exe
echo "正在编译win32.."
CGO_ENABLED=0 GOOS=windows GOARCH=386 go build  -o release/builds/win32/biu.exe
echo "正在编译darwin64.."
CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build  -o release/builds/darwin64/biu
echo "正在编译linux64.."
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build  -o release/builds/linux64/biu

echo "创建文件.."
mkdir release/apps/biu-win-32/
mkdir release/apps/biu-win-64/
mkdir release/apps/biu-darwin-64/
mkdir release/apps/biu-linux-64/
echo "复制相关文件.."
cp  release/common/win/* release/apps/biu-win-32/
cp  release/builds/win32/biu.exe release/apps/biu-win-32/biu.exe

cp  release/common/win/* release/apps/biu-win-64/
cp  release/builds/win64/biu.exe release/apps/biu-win-64/biu.exe

cp  release/builds/darwin64/biu release/apps/biu-darwin-64/biu

cp  release/builds/linux64/biu release/apps/biu-linux-64/biu
version="1.1"
echo "正在打包.."
zip -qj release/apps/biu-${version}-win32.zip release/apps/biu-win-32/*
zip -qj release/apps/biu-${version}-win64.zip release/apps/biu-win-64/*
zip -qj release/apps/biu-${version}-darwin64.zip release/apps/biu-darwin-64/*
zip -qj release/apps/biu-${version}-linux64.zip release/apps/biu-linux-64/*

echo "清除临时文件.."
rm -rf release/apps/biu-win-32/
rm -rf release/apps/biu-win-64/
rm -rf release/apps/biu-darwin-64/
rm -rf release/apps/biu-linux-64/
