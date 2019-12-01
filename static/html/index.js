// ====================== //
// Business Logic Helpers //
// ====================== //

function User(name, number, checked = false) {
    this.name = name;
    this.number = number;
    this.checked = checked;
}

User.prototype.getNumber = function() {
    return this.number;
}

User.prototype.isChecked = function() {
    return this.checked;
}

User.prototype.getName = function() {
    return this.name;
}

function getLabelForTextArea(users) {
    let result = "";
    if (users.length > 0) {
    result += "To: "
    }
    
    users.forEach(v => {
    if (v.isChecked()) {
        result += `${v.getName()}, `;
    }
    });

    return result;
}

// ============ //
// ERROR HELPER //
// ============ //

function ErrObj(message, state = null) {
    this.message = message;
    this.state = state;
}

ErrObj.prototype.log = function() {
    console.error(`message: ${this.message} \n state: ${this.state} \n`);
    return this;
}

ErrObj.prototype.throw = function() {
    throw this.message;
}

ErrObj.prototype.message = function() {
    return `${this.message}`;
}