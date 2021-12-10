const settings = {
    "async": true,
    "crossDomain": true,
    "url": "https://imdb8.p.rapidapi.com/auto-complete?q=the%20hangover",
    "method": "GET",
    "headers": {
        "x-rapidapi-host": "imdb8.p.rapidapi.com",
        "x-rapidapi-key": "1608c55388msh6143a2512f51131p168cc8jsnfa6987c5ea6a"
    }
};

$(document).ready(function () {
    $.getJSON(settings, function (result) {
        let text = "";
        var container = document.getElementById('imageContainer');
        var x = 0
        for (let i = 0; i < result['d'].length; i++) {
            if (result['d'][i]['q'] === "TV series" || result['d'][i]['q'] === "feature") {
                var element = document.createElement("div");
                var img = document.createElement('img');
                img.src = result['d'][i]['i']["imageUrl"];

                // Create a title under image
                const para = document.createElement("p");
                para.innerText = result['d'][i]['l']
                container.appendChild(para);

                // Append image and title in div
                container.appendChild(element).appendChild(img);
                container.appendChild(element).appendChild(para);
            }
        }
    });
});