# Web app using Microservices in Go: Part 2 - Implementation

## Introduction

In this part we will implement part of the microservices needed for our web app. We will implement the:
* Database
* a

## The database

### Design

After thinking through the design, I decided that it would be better if the database generated the task *Id*'s. This will also make it easier to get the last non-finished task and generate consecutive *Id*'s

How it will work:
* It will save new tasks assigning consecutive *Id*'s.
* It will remember the oldest not finished task.
* It will allow to get the last not finished task.
* It will allow to get the last not started task.
* It will allow to get a task by *Id*.
* It will allow to set a task by *Id*.
* The state will be represented by an int:
  * 0 - not started
  * 1 - in progress
  * 2 - finished
* Change the state to *not started* if it's been too long *in progress*. (maybe someone started to work but crashed)

![Database microservice post](https://www.lucidchart.com/publicSegments/view/4cf0690e-3dbb-42d9-befd-4a6efaaf6f72/image.png)

### Implementation

First, we should create the API and later we will add the implementations.
