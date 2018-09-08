# API não oficial dos autocarros da STCP

Este código em Go permite obter o tempo real de espra dos autocarros da
STCP no Porto, sob a forma de um leve serviço HTTP.

Uma vez que assenta sob um API não oficial, é possível que deixe
de funcionar a qualquer momento... Para que se possa obter um API
oficial, por favor apelem à câmara do Porto para que este seja
disponibilizado para livre uso dos programadores!


# Uso

```go
http.ListenAndServe(":8080", stcpbusapi.Handler)
```
