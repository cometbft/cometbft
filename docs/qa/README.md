---
order: 1
parent:
  title: CometBFT Quality Assurance
  description: This is a report on the process followed and results obtained when running v0.34.x on testnets
  order: 2
---

# CometBFT Quality Assurance

This directory keeps track of the process followed by the CometBFT team
for Quality Assurance before cutting a release.
This directory is to live in multiple branches. On each release branch,
the contents of this directory reflect the status of the process
at the time the Quality Assurance process was applied for that release.

File [method](./method.md) keeps track of the process followed to obtain the results
used to decide if a release is passing the Quality Assurance process.
The results obtained in each release are stored in their own directory.
The following releases have undergone the Quality Assurance process:

* [TM v0.34.x](./tm_v034/), which was tested just before releasing Tendermint Core v0.34.22
* [v0.34.x](./v034/), which was tested just before releasing v0.34.27
* [v0.37.x](./v037/), with TM v.34.x acting as a baseline
