### REMIND ME APP
```
  RemindMe is small web applications that stores reminder messages with a time to remind.

  When the time comes around to send the message (time to remind expires) RemindMe does so via sms.
```

### Current Versions with date:
- [ ] - `v1` : `Oct, 6, 2019`.

## Technologies - Descriptions
- [ ] Redis - Expire Mechanism
- [ ] Twilio API - sms
- [ ] NGINX - Web Template Server
- [ ] VueJS - Small Template embedded in html
- [ ] Docker - so far used for redis and nginx  
- [ ] Make - commands to control app: build, run, test, etc.


### Short Term Goals -
- [x] Test app live
- [x] Set Up Logging
- [x] Handle Errors
- [ ] Garbage handler for db expired calls past due
- [x] New sms features.
    - [x] Backend receives sms and handle by wiki bot.
- [] App for recipes.
 
- [ ] Frontend will have routing between apps (RemindMe & MealPlanner).


### Controlling the App 
  - [ ] `make run` - runs our sms application, if this fails we need to check the configs.
  - [ ] `make serve_static` - Spins up Web Server. Will find out what port is occupying `:80` and kill it if true before allocating it.

