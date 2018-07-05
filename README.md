# Tessellate


Exposes a GRPC (and a HTTP wrapper) around terraform to address:
* Remote and Reliable execution.
* Shared workspaces that don't need. provider credentials sharing.
* Searchable state.
* Authentication backends.
* First class ACL support.
* Watches for state change.
* Consul backed storage.
* Versioned history with Rollback.
* Hierarchical Variables.
* Asynchronous Jobs with abort.

## Data model

### Workspace
Workspace is the top most component of tessellate. It has a name and an optional map of variables that will be inherited by layouts when they are applied.


Operations
* Save
* Get
* GetLayouts

### Layout
Layout belongs to a workspace and has a name, an optional set of variables and a nested map of 
name and JSON template.
* This nested map is used to overcome terraform limitation of HCL which can be addressed when 0.12 will be released.
* Templates can be HCL compatible JSON or django styled template that yield a HCL compatible JSON.


Operations:
* Get
* Save
* Apply
* Destroy


*Asynchronous Jobs*

Both Apply and Destroy are long running operations and prone to failure. They are scheduled over an internal queue to be tried out by workers. Both the endpoints return a job-id that can be used to interact with the jobs.


*Priority* 

Tessellate has two queues per layout: apply and destroy. 

While apply jobs are queued and executed serially, the job dispatcher at end of a job run looks for out for any pending destroy command and prefers that over the remaining apply jobs.

Destroy has preference over other jobs because jobs are asynchronous and can result in backlog of pending updates, it might leave a destroy to be executed at the far end. If you end up accidentally scheduling a destroy, please use abort Job to abort the pending job.


### Variables

Variables are inherited by all children resources and can be overridden by child variables. Example: if a workspace was created with x=1, y=2 and a layout was created under workspace with vars z=3, p=5, y=4. And an apply call was made with t=7,z=7 the final set of Vars to be used would be:
X=1,y=4,p=5,z=7,z=7


*Use case*

A system manager wants to abstract the provider keys that should be used for a layout. Those can be seeded in the default workspace called staging. These variables are inherited by every layout run, the variable will be automatically merged into the apply vars. However, if a user wanted to enforce use of a specific account they could override the variable by supplying their own.

### Job

Job is a tessellate job that is long running terraform operation and streams out stdout & stderr logs when using streaming compatible clients.
(Please refer to grpc2 streaming clients for more information)


Jobs are picked from an internal queue and retried 5-times with exponential-backoff algorithm. If a job fails to run, it will be marked as a permanent failure.


Operations:
* Get
* Abort


Abort 

Aborts a job which is either in pending or failure state.

### Watch

Watch accepts a layoutId and a successful and/or a failure callback URL.
* On successful apply, a POST call is issued to the URL with {“state”: latest state, “job”: id}
* On a failure run, a POST call is issued with {“job”: id}


Watch is useful for building automatic moving job coordinations.

Callbacks use not-exactly-once delivery so there are situations where you may receive the callback more than once. You can implement idempotency using job id that is a part of the POST body.

Callbacks are tried for a maximum of 100 attempts, with linear-back-off. In case your receiver fails to be available during that time you will miss the callback.
One way to avoid this is to issue an LayoutApply again, once your callback receiver is up-and-running, and a no-op would end up triggering the callback again. Simple.


Operations:
* Start
* Stop


Start is to start watching a layout. And stop is to stop a watch.


Please note that only one watch is supported per layout. This will be improved later.

## ACL
* All requests are identified by a canonical identifier which is shipped as x_auth_id. Canonical id is of the form X.Y.Z where x is the first level group. Y is second level group and z is an id.
* By default, All resources created by x.y.156 are allowed for x.*.*
* The owner can then alter the rules and restrict it to any fine grained policy
* Policies only follow allow only model. Negations or deny are not supported at this moment.
* The creator of a resource has full access, forever, regardless of the allow policy. Example if x.y.2 had created the workspace, any layout with allow set to x.z.4 can still be controlled by x.y.2 
* This is designed to avoid lock-outs and avoid users creating resources on each other's behalf.


Alternatively, Casbin.org could be used.
