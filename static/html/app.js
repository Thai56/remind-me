  // ===================== //
  // Date and Time Helpers //
  // ===================== //

  function formatTime(date) {
    const d = new Date(date); 
    const hours = d.getHours();
    const minutes = d.getMinutes();
    const seconds = d.getSeconds();

    return `${hours < 10 ? `0${hours}` : hours}:${minutes < 10 ? `0${minutes}` : minutes}:${seconds < 10 ? `0${seconds}` : seconds}`
}

function formatDate(date) {
    let d = new Date(date),
        month = '' + (d.getMonth() + 1),
        day = '' + d.getDate(),
        year = d.getFullYear();

    if (month.length < 2) 
        month = '0' + month;
    if (day.length < 2) 
        day = '0' + day;

    return [year, month, day].join('-');
}

function getAsDate(day, time) {
    var hours = Number(time.match(/^(\d+)/)[1]);
    var minutes = Number(time.match(/:(\d+)/)[1]);
    var AMPM = time.match(/\s(.*)$/)[1];
    if(AMPM == "pm" && hours<12) hours = hours+12;
    if(AMPM == "am" && hours==12) hours = hours-12;
    var sHours = hours.toString();
    var sMinutes = minutes.toString();
    if(hours<10) sHours = "0" + sHours;
    if(minutes<10) sMinutes = "0" + sMinutes;
    time = sHours + ":" + sMinutes + ":00";
    var d = new Date(day);
    var n = d.toISOString().substring(0,10);
    var newDate = new Date(n+"T"+time);
    console.log("returning new date", newDate);
    return newDate;
}

function formatAMPM(hours, minutes) {
    var ampm = hours >= 12 ? 'pm' : 'am';
    hours = hours % 12;
    hours = hours ? hours : 12; // the hour '0' should be '12'
    minutes = minutes < 10 ? '0'+minutes : minutes;
    var strTime = hours + ':' + minutes + ' ' + ampm;
    return strTime;
}

function formatMMDDYYY(date) {
    date = date.split("-")
    const month = date[1];
    const year = date[0];
    const day = date[2];

    return `${month}/${day}/${year}`;
}

// ============== //
// Default States //
// ============== //
const stateDefaults = {
    getReminderDefaults() {
        return {
        content: "",
        users: [
            new User("Thai", "7073447433", true),
            new User("Chipi", "7073866800"),
        ],
        errWithSend: "",
        sendErr: false,
        responseSaved: false,
        valid: false,
        };
    },
    getUserDefaults() {
        return {
        };
    },
    getDatePickerDefaults() {
        return {
        nowMenu: false,
        currentDate: formatDate(new Date()),
        };
    },
    getTimePickerDefaults() {
        return {
        currentTime: formatTime(new Date()),
        timeDialog: false,
        };
    },
    getTextAreaDefaults() {
        return {
        counter: 0,
        textAreaLabel: getLabelForTextArea(this.getReminderDefaults().users),
        textAreaLoading: false,
        placeholder: 'do the dishes...',
        maxTextLength: 40,
        rowHeight: 24,
        rows: 1,
        };
    },
    getLayoutDefaults() {
        return {
        drawers: ['Default (no property)', 'Permanent', 'Temporary'],
        primaryDrawer: {
            model: null,
            type: 'default (no property)',
            clipped: false,
            floating: false,
            mini: false,
        }
        };
    },
    getFooterDefaults() {
        return {
        footer: {
            inset: false,
        },
        };
    },
}

  // ================= //
  // Vue Logic Helpers //
  // ================= //
  new Vue({
    el: '#app',
    vuetify: new Vuetify({
      theme: {
        dark: true,
      },
    }),
    data: () => ({
      ...stateDefaults.getReminderDefaults(),
      ...stateDefaults.getUserDefaults(),
      ...stateDefaults.getDatePickerDefaults(),
      ...stateDefaults.getTimePickerDefaults(),
      ...stateDefaults.getTextAreaDefaults(),
      ...stateDefaults.getLayoutDefaults(),
      ...stateDefaults.getFooterDefaults(),
    }),
    methods: {
      sendReminder: async function (event) {
        // trigger loading for text area
        this.textAreaLoading = true;  
        
        event.preventDefault();
        try {
            const body = {
                name: "Username",
                destinations: this.getDestinations(),
                content: this.getContent(),
                expire: this.getUIDateTime(),
            };
            const request = {
                method: 'POST',
                mode: 'cors',
                headers: {
                    'Content-Type': 'application/json',
                    'Access-Control-Allow-Origin': '*',
                },  
                body: JSON.stringify(body),
            };
            console.debug("Sending reminder body : ", body);
            const response = await fetch('http://localhost:8080/remind', request);
            
            // trigger loading finished for text area
            this.textAreaLoading = false;
            
            const res = await response.json();
            if (res.error != null) {
              new ErrObj(res.error.message, res.error.code)
                .log().throw()
            }
            console.log("res", JSON.stringify(res));
            
            this.triggerSuccess();
        } catch (err){
            this.textAreaLoading = false;
            this.triggerError(err);
        }  
      },
      triggerError: function (err) {
          this.sendErr = true;
          this.errWithSend = `${err}`;
      },
      closeErrorDialog: function() {
        this.sendErr = false;
        this.errWithSend = "";
      },
      triggerSuccess: function () {
          this.responseSaved = true;
      },
      closeSuccessDialog: function() {
        this.responseSaved = false
      },
      resetState: function() {
          const defaultState = stateDefaults.getReminderDefaults()
          for (key in defaultState) {
              this[key] = defaultState[key];
          }
      },
      getUIDateTime: function() {
          const timeData = this.currentTime.slice().split(":");
          const hours = timeData[0];
          const minutes = timeData[1];

          const currentDate = this.currentDate.slice();
          const formattedDate = formatMMDDYYY(currentDate);
          const date = getAsDate(formattedDate, formatAMPM(hours, minutes));

          if (date.getTime() < new Date().getTime()) {
              return new ErrObj(`Time has been surpassed ${new Date(date.getTime())}`, this.currentTime)
                  .log()
                  .throw();
          }

          return date.getTime();
      },
      getDestinations: function() {
        const phoneNumbers = this.users.reduce((acc, v) => {
          if (v.isChecked()) {
            acc.push(v.getNumber());
          }

          return acc;
        }, []);

        if (phoneNumbers.length === 0) {
            return new ErrObj(
                'Failed precondition. \nPlease check the destinations phone number.',
                phoneNumbers
            ).log().throw();
        }

        return phoneNumbers;
      },
      getContent: function() {
          if (!this.content) {
              return new ErrObj(
                  'Failed precondition. \nPlease check the message content textbox.',
                  this.content
              ).log().throw();
          }

          return this.content;
      },
      updateTextAreaLabel() {
        this.textAreaLabel = getLabelForTextArea(this.users)
      },
    },
  })
