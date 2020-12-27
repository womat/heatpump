set GOARCH=arm
set GOOS=linux
go build -o ..\bin\heatpump ..\cmd\heatpump.go

set GOARCH=386
set GOOS=windows
go build -o ..\bin\heatpump.exe ..\cmd\heatpump.go