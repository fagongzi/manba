function getChoose(data, attr) {
    var choosees = [];
    var len = data.length;

    for (var i = 0; i < len; i++) {
        if (data[i].checked) {
            choosees[choosees.length] = data[i][attr];
        }
    }

    return choosees;
}