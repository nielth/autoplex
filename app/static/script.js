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


$(document).ready(function () {
    const currentPage = window.location.pathname;
    if (currentPage.includes("/movie/")){
        $("input[name='movie']").attr('checked', 'checked');
    }
    else if (currentPage.includes("/series/")){
        $("input[name='series']").attr('checked', 'checked');
    }
});


document.getElementById("titleButton").addEventListener("click", function () {

    const inputVal_space = document.getElementById("title").value.replace(" ", "%20");
    const inputVal = inputVal_space.replace(" ", "%20");

    if (document.getElementById("movie").checked) {
        location.href = `/search/movie/${inputVal}`;
    } else if (document.getElementById("series").checked) {
        location.href = `/search/series/${inputVal}`;
    } else {
        alert("Choose movie or tv series");
    }
})