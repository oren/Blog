# Web app using Microservices in Go: Part 1 - Design

## Introduction

Recently it's a constantly repeated buzzword - ***Microservices***. *You can love 'em or hate 'em, but you really shouldn't ignore 'em*. In this short series we'll create a web app using a microservice architecture. We'll try not to use 3rd party tools and libraries. Remember though that when creating a production web app it is highly recommended to use 3rd party libraries (even if only to save you time).

We will create the various components in a basic form. We won't use advanced *caching* or use a *database*. We will create a basic **key-value** store and a simple storage service. We will use the ***Go*** language for all this.

## The functionality

First we should decide what our web app will do. The web app we'll create in this series will get an image from a user and give back an unique **ID**. The image will get modified using complicated and highly sophisticated algorithms, like swapping the blue and red channel, and the user will be able to use the *ID* to check if the work on the image has been finished already or if it's still in progress. If it's finished he will be able to download the altered image.

## Designing the architecture

We want the **architecture** to be microservices, so we should design it like that. We'll for sure need a service facing the user, the one that provides the *interface* for communication with our app. This could also handle *authentication*, and should be used as the service redirecting the workload to the right sub-services. (useful if you plan to integrate more funcionality into the app)

We will also want a microservice which will handle all our images. It will get the image, generate an *ID*, store information related to each *task*, and save the images. To handle high workloads it's a good idea to use a ***master-slave*** system for our image modification service. The image handler will be the *master*, and we will create *slave* microservices which will ask the *master* for images to work on.

We will also need a *key-value* datastore for various *configuration*, a storage system, for saving our images, pre- and post-modification, and a database-ish service holding the information about each task.

This should suffice to begin with.

Here I'd like to also state that the architecture could change during the series if needed. And I encourage you to comment if you think that something could be done better.

### Communication

We will also need to define the method the services communicate by. In this app we will use ***REST*** everywhere. You could also use a ***message BUS*** or ***Remote Procedure Calls*** - short ***RPC***, but I won't write about them here.

### Designing the microservice API's

Another important thing is to design the ***API***'s of you microservices. We will now design each of them to get an understanding about what they are for.

#### The key-value store

This one's mainly for configuration. It will have a simple post-get interface:

* POST:
	* Arguments:
		* Key
		* Value
	* Response:
		* Success/Failure
* GET:
	* Arguments:
		* Key
	* Response:
		* Value/Failure

#### The storage

Here we will store the images, again using a key-value interface and an argument stating if this one's pre- or post-modification. For the sake of simplicity we will just save the image to a folder named, depending on the state of the image, finished/inProgress.

* POST:
	* Arguments:
		* Key
		* State: pre-/post-modification
		* Data
	* Response:
		* Success/Failure
* GET:
	* Arguments:
		* Key
		* State: pre-/post-modification
	* Response:
		* Data/Failure

#### Database

This one will save our tasks. If they are waiting to start, in progress or finished, their Id.

* POST:
	* Arguments:
		* TaskId
		* State: not started/ in progress/ finished
	* Response:
		* Success/Failure
* GET:
	* Arguments:
		* TaskId
	* Response:
		* State/Failure
* GET:
	* Path:
		* not started/ in progress/ finished
	* Reponse:
		* list of TaskId's

#### The Frontend

The frontend is there mainly to provide a communication way between the various services and the user. It can also be used for authentication and authorization.

* POST:
	* Path:
		* newImage
	* Arguments:
		* Data
	* Response:
		* Id
* GET:
	* Path:
		* image/isReady
	* Arguments:
		* Id
	* Response:
		* not found/ in progress / finished
* GET:
	* Path:
		* image/get
	* Arguments:
		* Id
	* Response:
		* Data

#### Image master microservice

This one will get new images from the fronted/user and send them to the storage service. It will also create a new task in the database, and orchestrate the workers who can ask for work and notify when it's finished.

* Frontend interface:
	* POST:
		* Path:
			* newImage
		* Arguments:
			* Data
		* Response:
			* Id
	* GET:
		* Path:
			* isReady
		* Arguments:
			* Id
		* Response:
			* not found/ in progress / finished
	* GET:
		* Path:
			* get
		* Arguments:
			* Id
		* Response:
			* Data/Failure
* Worker interface:
	* GET:
		* Path:
			* getWork
		* Response:
			* Id/noWorkToDo
	* POST:
		* Path:
			* workFinished
		* Arguments:
			* Id
		* Response:
			* Success/Failure

#### Image worker microservice

This one doesn't have any API. It is a client to the master image service, which he finds using the key-value store. He gets the image data to work on from the storage service.

## Scheme



## Conclusion

This is basically everything regarding the design. In the next part we will write part of the microservices. Again, I encourage you to comment expressing what you think about this design!
