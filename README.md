# conpass
Console password manager


## How it uses?

1) Set the master password

2) Use it
 - Add the new data: conpass add -n some-name -p some-password  
   -n, --name: adding data name
   -data, --data: adding data
   
 - Get data: conpass get -n some-name  
   -b, --buffer: copy to buffer
   
 - Update data: conpass edit -n some-name -p new-password
 - Remove data: conpass rm -n some-name
   

