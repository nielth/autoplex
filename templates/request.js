const settings = {
    "async": true,
    "crossDomain": true,
    "url": "",
    "method": "GET",
    "headers": {
        "x-rapidapi-host": "imdb8.p.rapidapi.com",
        "x-rapidapi-key": "1608c55388msh6143a2512f51131p168cc8jsnfa6987c5ea6a"
    }
};

// Need fix
$("#title").keyup(function(event) {
    if (event.keyCode === 13) {
        $("#titleButton").click();
    }
});

function getTitleValue() {
    const container = document.getElementById('imageContainer');
    // Remove previous search
    container.textContent = '';
    // Selecting the input element and get its value
    const inputVal = document.getElementById("title").value;

    // Removes unnecessary spaces
    var newStr = inputVal.replaceAll(" ", "%");
    while (newStr[newStr.length - 1] === "%") {
        newStr = newStr.slice(0, -1);
    }

    settings.url = `https://imdb8.p.rapidapi.com/auto-complete?q=${newStr}`;

    $.getJSON(settings, function (result) {
        for (let i = 0; i < result['d'].length; i++) {
            if (result['d'][i]['q'] === "TV series" || result['d'][i]['q'] === "feature") {
                const element = document.createElement("div");

                // Find the image
                const img = document.createElement('img');
                img.src = result['d'][i]['i']["imageUrl"];
                img.alt = result['d'][i]['id']
                img.onclick = function () {
                    titleValue(event)
                };

                // Find the title
                const para = document.createElement("p");
                para.innerText = result['d'][i]['l'];
                container.appendChild(para);

                // Append image and title into div
                container.appendChild(element).appendChild(img);
                container.appendChild(element).appendChild(para);
            }
        }
    });
}

// Need fix, can't read JSON
function titleValue(event) {
    const container = document.getElementById('imageContainer');
    container.textContent = '';

    const alt = event.target.alt;

    $.ajax({url: `http://127.0.0.1:5000/${alt}/`, success: function(result){
      alert(result);
    }});
}