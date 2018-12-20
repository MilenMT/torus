# Torus

## Torus Overview

Torus is an open source project for distributed storage coordinated through 

Torus provides a resource pool and basic file primitives from a set of daemons running atop multiple nodes. These primitives are made consistent by being append-only and coordinated by [etcd]. From these primitives, a Torus server can support multiple types of volumes, the semantics of which can be broken into subprojects. It ships with a simple block-device volume plugin, but is extensible to more.


Sharding is done via a consistent hash function, controlled in the simple case by a hash ring algorithm, but fully extensible to arbitrary maps, rack-awareness, and other nice features. The project name comes from this: a hash 'ring' plus a 'volume' is a torus. 

## Project Status and Background

This project is experimental status at this point.

After forked from [original project] (retired currently), this project integrated


It plans to create a bit more new feature, but performance improvment is the highest priority.

## Trying out Torus

To get started quicky using Torus for the first time, start with the guide to , learn more about setting up Torus on Kubernetes using FlexVolumes 

## Contributing to Torus

Torus is an open source project and contributors are welcome!

## Licensing

Unless otherwise noted, all code in the Torus repository is licensed under the [Apache 2.0 license](LICENSE). Some portions of the codebase are derived from other projects under different licenses; the appropriate information can be found in the header of those source files, as applicable.
