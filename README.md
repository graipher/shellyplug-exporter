# ShellyPlug Exporter

Periodically get information from a [Shelly Plus Plug S](https://www.shelly.com/de/products/shop/shelly-plus-plug-s-1) and publish it as Prometheus metrics.

## Exposed metrics

Output button state, power, voltage, current, total energy, temperature (in °C), updates available and last update time.

# How to run

Build and run with Docker:

```sh
docker build -t shellyplug-exporter .
docker run --it -rm \
    -e SHELLYPLUG_URL=host/ip_address \
    shellyplug-exporter
```

## Limitations

* the port the API listens on is currently hard-coded to 2112
* the period with which the ShellyPlug is scraped is hard-coded to 1 minute
