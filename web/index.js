// global vars
const input = document.getElementById('input');
const preview = document.getElementById('preview');
const notification = document.getElementById('notification');
const app = document.getElementById('app');
let timeout;

// dark mode
let darkmode = localStorage.getItem('darkmode') === '1';
if (darkmode) {
    app.classList.add('dark-mode');
}


// preview mode
// it's the inverse of edit, so it is automatically true
// if the edit item doesn't exist in localStorage
let previewState = localStorage.getItem('edit') !== '1';
if (previewState) {
    app.classList.add('preview')
}

// on initial page load
preview.innerHTML = marked(input.value);

function render() {
    preview.innerHTML = marked(input.value);

    // simple timeout function, so that the documents doesn't get
    // saved on every keystroke, but rather 500ms after the user has
    // stopped typing
    clearTimeout(timeout);
    timeout = setTimeout(_ => {
        fetch(window.location, {
            method: 'PUT',
            body: input.value
        }).catch(err => {
            setNotification('Error: ' + err)
        })
    }, 500)
}

// simple notification, used to alert the user if the document
// couldn't get saved. For now it has the fixed color red
function setNotification(text) {
    notification.innerText = text;
    notification.style.display = 'block';
    setTimeout(_ => notification.style.display = 'none', 2000);
}

// file upload returns the upladed element id
function addAttachment(file) {
    const data = new FormData()
    data.append('file', file)

    fetch('/attachments', {
        method: 'POST',
        headers: {
            'accept': 'application/json'
        },
        body: data
    }).then(res => res.json()).then(res => {
        let prefix = ""
        if (file.type.startsWith('image/')) {
            prefix = "!"
        }

        insertAtCursor(input,
            `${prefix}[${file.name}](${res.file})`)
        render()
    }).catch(err => setNotification('Error: ' + err))
}

// on clipboard drop
input.ondrop = (e) => {
    e.preventDefault();
    addAttachment(e.dataTransfer.files[0])
}

// paste event
input.addEventListener('paste', event => {
    let items = (event.clipboardData || window.clipboardData).items;

    for (const index in items) {
        const item = items[index]

        if (item.kind !== 'file') {
            return
        }

        addAttachment(item.getAsFile())
    }
})

function insertAtCursor(element, value) {
    if (element.selectionStart || element.selectionStart == '0') {
        var startPos = element.selectionStart;
        var endPos = element.selectionEnd;
        element.value = element.value.substring(0, startPos)
            + value
            + element.value.substring(endPos, element.value.length);
    } else {
        element.value += value;
    }
}



// shortcuts
document.onkeydown = (e) => {
    if (e.keyCode == 68 && (e.ctrlKey || e.metaKey) && e.shiftKey) {    // CTRL + SHIFT + D
        darkmode = !darkmode;
        localStorage.setItem("darkmode", darkmode ? "1" : "0");

        if (darkmode) {
            app.classList.add('dark-mode')
        } else {
            app.classList.remove('dark-mode')
        }

        e.preventDefault();
        return false;
    } else if (e.keyCode == 68 && (e.ctrlKey || e.metaKey) && !e.shiftKey) { // CTRL + D
        previewState = !previewState;
        localStorage.setItem("edit", previewState ? "0" : "1");

        if (previewState) {
            app.classList.add('preview')
        } else {
            app.classList.remove('preview')
        }

        e.preventDefault();
        return false
    }
}