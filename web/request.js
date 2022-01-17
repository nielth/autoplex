/*
#############################
#                           #
#   WARNING! I suck at JS   #
#                           #
#############################
*/


$("#title").keyup(function (event) {
    if (event.keyCode === 13) {
        $("#titleButton").click();
    }
});

$('input[type="checkbox"]').on('change', function () {
    $('input[type="checkbox"]').not(this).prop('checked', false);
});

$(document).ready(function() {
    $("input[name='movie']").attr('checked', 'checked');
});


document.getElementById("titleButton").addEventListener("click", function () {    

    const inputVal = document.getElementById("title").value;

    if(document.getElementById("movie").checked){
        location.href = `/search/movie/${inputVal}`;
    } else if(document.getElementById("series").checked){
        location.href = `/search/series/${inputVal}`;
    } else {
        alert("Choose movie or tv series");
    }
})