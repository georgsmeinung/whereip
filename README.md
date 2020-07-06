## whereip
# Welcome to Where IP!
REST service to collect geographical data for a given IP. This an app that gets the Country for a given IP and related information. It supports both IPv4 and IPv6

You can use it as a REST service in a line like:

```
AU GET {sever_ip}:3000/whereip/1.1.1.1
US GET {sever_ip}:3000/whereip/2606:4700:4700::1111
BR GET {sever_ip}:3000/whereip/200.223.129.162
ES GET {sever_ip}:3000/whereip/195.53.69.132
IN GET {sever_ip}:3000/whereip/203.115.71.66
```

And you will get and answer like this:
```JSON
{
    "From": "1.1.1.1",
    "When": "2020-03-26T11:30:56-0300",
    "CountryCode": "AU",
    "CountryName": "Australia",
    "Languages": [
        "English"
    ],
    "Timezones": [
        "2020-03-26T14:30:56+0000",
        "2020-03-26T14:33:56+0003",
        "2020-03-26T14:30:56+0000",
        "2020-03-26T14:30:56+0000",
        "2020-03-26T14:33:56+0003",
        "2020-03-26T15:30:56+0100",
        "2020-03-26T15:33:56+0103",
        "2020-03-26T15:33:56+0103"
    ],
    "Distance": 13076,
    "Currency": "AUD",
    "ExRate": 0.83484
}
```

You can get view and clear stats from here:
```
- GET {sever_ip}:3000/stats/
- GET {sever_ip}:3000/fullstats/
- GET {sever_ip}:3000/clearstats/
```
