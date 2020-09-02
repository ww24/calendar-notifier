你好！
很冒昧用这样的方式来和你沟通，如有打扰请忽略我的提交哈。我是光年实验室（gnlab.com）的HR，在招Golang开发工程师，我们是一个技术型团队，技术氛围非常好。全职和兼职都可以，不过最好是全职，工作地点杭州。
我们公司是做流量增长的，Golang负责开发SAAS平台的应用，我们做的很多应用是全新的，工作非常有挑战也很有意思，是国内很多大厂的顾问。
如果有兴趣的话加我微信：13515810775  ，也可以访问 https://gnlab.com/，联系客服转发给HR。
calendar-notifier
===

![Test on master][github-actions-img]

Calendar Notifier provides event handler and actions triggered by Google Calendar.

[![dockeri.co][dockeri-img]][dockeri-url]

## Features

- Events
  - [x] [Calendar Events List API](https://developers.google.com/calendar/v3/reference/events/list)
  - [ ] [Calendar Events Watch API](https://developers.google.com/calendar/v3/reference/events/watch)
- Actions
  - [x] HTTP Action
  - [x] [Cloud Pub/Sub](https://cloud.google.com/pubsub/) Action
  - [x] [Cloud Tasks](https://cloud.google.com/tasks/) Action

## Setup

- Create GCP Project
  - https://cloud.google.com/resource-manager/docs/creating-managing-projects
- Create GCP Service Account
  - Role assignment is not required if it does not use pubsub action.
  - https://cloud.google.com/iam/docs/creating-managing-service-accounts
- Share Google Calendar with service account
  - Use service account e-mail address. (e.g. `my-service-account@project-id.iam.gserviceaccount.com`)
  - https://support.google.com/calendar/answer/37082
- Set `SERVICE_ACCOUNT` environment. Use base64 encoded GCP Service Account JSON.
  - `echo "SERVICE_ACCOUNT=$(base64 < service_account.json)" > .env`
- Edit `config.sample.yml` and save as `config.yml`.
- Set `CONFIG` environment. Use base64 encoded config.yml.
  - `echo "CONFIG=$(base64 < config.yml)" > .env`

## Usage

### Use docker-compose

- Run `docker-compose up`

### Use docker

- Run `docker build -t calendar-notifier .`
- Run `docker run -e SERVICE_ACCOUNT=$(base64 < service_account.json) -e CONFIG=$(base64 < config.yml) calendar-notifier`

## Permission

### Cloud Tasks

If use Cloud Tasks Action, service account should have the following permissions.

```
cloudtasks.tasks.list
cloudtasks.tasks.create
cloudtasks.tasks.delete
iam.serviceAccounts.actAs
```

For example, give your service account the following roles.

```
roles/cloudtasks.viewer
roles/cloudtasks.enqueuer
roles/cloudtasks.taskDeleter
roles/iam.serviceAccountUser
```

#### References
- https://cloud.google.com/tasks/docs/reference-access-control
- https://cloud.google.com/iam/docs/understanding-service-accounts?hl=ja#sa_common


### Cloud Pub/Sub

If use Cloud Pub/Sub Action, service account should have the following permissions.

```
pubsub.topics.publish
```

For example, give your service account the following roles.

```
roles/pubsub.publisher
```


#### References
- https://cloud.google.com/pubsub/docs/access-control

[github-actions-img]: https://github.com/ww24/calendar-notifier/workflows/Test%20on%20master/badge.svg?branch=master
[dockeri-img]: https://dockeri.co/image/ww24/calendar-notifier
[dockeri-url]: https://hub.docker.com/r/ww24/calendar-notifier
