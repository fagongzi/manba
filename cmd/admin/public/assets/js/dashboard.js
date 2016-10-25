function routeDashboard($routeProvider) {
    $routeProvider.when('/', {
        "templateUrl" : "html/dashboard.html",
        "controller" : DashboardController
    }).when('/analysis/:proxyAddr/:serverAddr/:secs', {
        "templateUrl" : "html/analysis/analysis.html",
        "controller" : AnalysisController
    });
}

function DashboardController($scope, $routeParams, $http) {
    // refreshCanvas($scope, $routeParams, $http)
}

function AnalysisController($scope, $routeParams, $http) {
    $scope.serverAddr = $routeParams.serverAddr
    refreshCanvas($scope, $routeParams, $http)
}

function refreshCanvas($scope, $routeParams, $http) {
    $scope.metrics = {};
    $scope.proxyAddr = $routeParams.proxyAddr;

    setInterval(function () {
        var proxyAddr = $routeParams.proxyAddr;
        var serverAddr = $routeParams.serverAddr;
        var secs = $routeParams.secs;

        $http.get("api/analysis/" + proxyAddr + "/" + serverAddr + "/" + secs).success(function (data) {
            $scope.metrics = data.value || {};
        });
    }, 1000);


    Highcharts.setOptions({
        global: {
            useUTC: false
        }
    });

    $('#containerCount').highcharts({
        chart: {
            type: 'spline',
            animation: Highcharts.svg, // don't animate in old IE
            marginRight: 10,
            events: {
                load: function () {
                    var requests = this.series[0];
                    var qps = this.series[1];
                    var failures = this.series[2];
                    var rejects = this.series[3];
                    var successed = this.series[4];
                    
                    setInterval(function () {
                        var x = (new Date()).getTime();
                        qps.addPoint([x, $scope.metrics.requestCount || 0], true, true);
                        requests.addPoint([x, $scope.metrics.qps || 0], true, true);
                        failures.addPoint([x, $scope.metrics.requestFailureCount || 0], true, true);
                        rejects.addPoint([x, $scope.metrics.rejectCount || 0], true, true);
                        successed.addPoint([x, $scope.metrics.requestSuccessedCount || 0], true, true);
                    }, 1000);
                }
            }
        },
        title: {
            text: $routeParams.serverAddr + ' Counts in lastest seconds'
        },
        xAxis: {
            type: 'datetime',
            tickPixelInterval: 300
        },
        yAxis: {
            title: {
                text: 'Counts'
            },
            plotLines: [{
                value: 0,
                width: 1,
                color: '#808080'
            }]
        },
        tooltip: {
            formatter: function () {
                return '<b>' + this.series.name + '</b><br/>' +
                    Highcharts.dateFormat('%Y-%m-%d %H:%M:%S', this.x) + '<br/>' +
                    Highcharts.numberFormat(this.y, 2);
            }
        },
        legend: {
            enabled: true
        },
        exporting: {
            enabled: false
        },
        series: [{
            name: 'requests',
            data: (function () {
                // generate an array of random data
                var data = [],
                    time = (new Date()).getTime(),
                    i;

                for (i = -20; i <= 0; i += 1) {
                    data.push({
                        x: time + i * 1000,
                        y: 0
                    });
                }
                return data;
            }())
        },{
            name: 'qps',
            data: (function () {
                // generate an array of random data
                var data = [],
                    time = (new Date()).getTime(),
                    i;

                for (i = -20; i <= 0; i += 1) {
                    data.push({
                        x: time + i * 1000,
                        y: 0
                    });
                }
                return data;
            }())
        },{
            name: 'failure',
            data: (function () {
                // generate an array of random data
                var data = [],
                    time = (new Date()).getTime(),
                    i;

                for (i = -20; i <= 0; i += 1) {
                    data.push({
                        x: time + i * 1000,
                        y: 0
                    });
                }
                return data;
            }())
        },{
            name: 'rejects',
            data: (function () {
                // generate an array of random data
                var data = [],
                    time = (new Date()).getTime(),
                    i;

                for (i = -20; i <= 0; i += 1) {
                    data.push({
                        x: time + i * 1000,
                        y: 0
                    });
                }
                return data;
            }())
        },{
            name: 'successed',
            data: (function () {
                // generate an array of random data
                var data = [],
                    time = (new Date()).getTime(),
                    i;

                for (i = -20; i <= 0; i += 1) {
                    data.push({
                        x: time + i * 1000,
                        y: 0
                    });
                }
                return data;
            }())
        }]
    });


    $('#containerDelays').highcharts({
        chart: {
            type: 'spline',
            animation: Highcharts.svg, // don't animate in old IE
            marginRight: 10,
            events: {
                load: function () {
                    var max = this.series[0];
                    var min = this.series[1];
                    var avg = this.series[2];
                    
                    setInterval(function () {
                        var x = (new Date()).getTime();
                        max.addPoint([x, $scope.metrics.max || 0], true, true);
                        min.addPoint([x, $scope.metrics.min || 0], true, true);
                        avg.addPoint([x, $scope.metrics.avg || 0], true, true);
                    }, 1000);
                }
            }
        },
        title: {
            text: $routeParams.serverAddr + ' Analysis RT in lastest seconds'
        },
        xAxis: {
            type: 'datetime',
            tickPixelInterval: 300
        },
        yAxis: {
            title: {
                text: 'millseconds'
            },
            plotLines: [{
                value: 0,
                width: 1,
                color: '#808080'
            }]
        },
        tooltip: {
            formatter: function () {
                return '<b>' + this.series.name + '</b><br/>' +
                    Highcharts.dateFormat('%Y-%m-%d %H:%M:%S', this.x) + '<br/>' +
                    Highcharts.numberFormat(this.y, 2);
            }
        },
        legend: {
            enabled: true
        },
        exporting: {
            enabled: false
        },
        series: [{
            name: 'max',
            data: (function () {
                // generate an array of random data
                var data = [],
                    time = (new Date()).getTime(),
                    i;

                for (i = -20; i <= 0; i += 1) {
                    data.push({
                        x: time + i * 1000,
                        y: 0
                    });
                }
                return data;
            }())
        },{
            name: 'min',
            data: (function () {
                // generate an array of random data
                var data = [],
                    time = (new Date()).getTime(),
                    i;

                for (i = -20; i <= 0; i += 1) {
                    data.push({
                        x: time + i * 1000,
                        y: 0
                    });
                }
                return data;
            }())
        },{
            name: 'avg',
            data: (function () {
                // generate an array of random data
                var data = [],
                    time = (new Date()).getTime(),
                    i;

                for (i = -20; i <= 0; i += 1) {
                    data.push({
                        x: time + i * 1000,
                        y: 0
                    });
                }
                return data;
            }())
        }]
    });
}