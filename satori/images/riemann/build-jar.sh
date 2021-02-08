#!/bin/bash

# require lein+jdk

lein uberjar && \
  cp target/riemann-*-satori-standalone.jar app/