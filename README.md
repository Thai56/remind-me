### REMIND ME APP
```
  RemindMe is small web applications that stores reminder messages with a time to remind.

  When the time comes around to send the message (time to remind expires) RemindMe does so via sms.
```

### Current Versions with date:
- [ ] - `v1` : `Oct, 6, 2019`.

## Technologies - Descriptions
- [ ] Redis - Expire Mechanism, Storage
- [ ] Twilio API - sms
- [ ] NGINX - Web Template Server
- [ ] VueJS - Small Template embedded in html
- [ ] Docker - so far used for redis and nginx  
- [ ] Make - commands to control app: build, run, test, etc.


### Short Term Goals -
- [ ] Test app live
    - [ ] Create Issues when arise.
- [ ] Set Up Logging
- [ ] Handle Errors
- [ ] Garbage handler for db expired calls past due


### Controlling the App 
  - [ ] `make run` - runs our sms application, if this fails we need to check the configs.
  - [ ] `make serve_static` - Spins up Web Server. Will find out what port is occupying `:80` and kill it if true before allocating it.

