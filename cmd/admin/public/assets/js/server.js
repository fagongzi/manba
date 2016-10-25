function routeServer($routeProvider) {
    $routeProvider.when("/servers", {
        "templateUrl": "html/server/list.html",
        "controller": ServerIndexController
    }).when("/servers/:name", {
        "templateUrl": "html/server/detail.html",
        "controller": ServerDetailController
    }).when("/new/server", {
        "templateUrl": "html/server/new.html",
        "controller": ServerCreateController
    });
}

function ServerCreateController($scope, $routeParams, $http, $location, $route) {
    $scope.add = function () {
        var c = {
            "schema": "http",
            "addr": $scope.newServerAddr,
            "checkPath": $scope.newServerCheckPath,
            "checkResponsedBody": $scope.newServerCheckResponsedBody,
            "checkTimeout": $scope.newServerCheckTimeout,
            "checkDuration": $scope.newServerCheckDuration,
            "maxQPS": $scope.newServerMaxQPS,
            "halfToOpen": $scope.newServerHalfToOpen,
            "halfTrafficRate": $scope.newHalfTrafficRate,
            "closeCount": $scope.newServerToCloseCount,
        };

        $http.post('api/servers', c).success(function (data) {
            $location.path("/servers");
            $route.reload();
        });
    }
}

function ServerIndexController($scope, $routeParams, $http, $location, $route) {
    $http.get("api/servers").success(function (data) {
        $scope.servers = data.value
    });

    $http.get("api/proxies").success(function (data) {
        $scope.proxies = data.value
    });

    $http.get("api/clusters").success(function (data) {
        $scope.clusters = data.value

        var len = $scope.clusters.length;
        for (var i = 0; i < len; i++) {
            $scope.clusters[i].checked = false;
        }
    });


    $scope.create = function () {
        $location.path("/new/server");
        $route.reload();
    }

    $scope.delete = function (addr) {
        $http.delete('api/servers/' + addr).success(function (data) {
            $location.path("/servers");
            $route.reload();
        });
    }

    $scope.preBind = function (addr) {
        $scope.bindServer = addr
    }

    $scope.bind = function () {
        var clusters = getChoose($scope.clusters, "name");

        if (clusters.length > 0) {
            b = {
                "clusterName": clusters[0],
                "serverAddr": $scope.bindServer
            };

            $http.post('api/binds', b).success(function (data) {
                $location.path("/servers");
                $route.reload();
            });
        }
    }

    $scope.addAnalysis = function (addr) {
        d = {
            "proxyAddr": $scope.proxyAddr,
            "serverAddr": addr,
            "secs": 1
        };

        $http.post('api/analysis', d).success(function (data) {
            $location.path("/servers");
            $route.reload();
        });
    }
    $scope.goAnalysis = function (addr) {
        $location.path("/analysis/" + $scope.proxyAddr + "/" + addr + "/1");
        $route.reload();
    }
}

function ServerDetailController($scope, $routeParams, $http, $location, $route) {
    $http.get("api/servers/" + $routeParams.name).success(function (data) {
        $scope.server = data.value;
    });

    $scope.unbind = function (svrAddr, clusterName) {
        b = {
            "clusterName": clusterName,
            "serverAddr": svrAddr
        };

        $http.delete('api/binds', { "data": b }).success(function (data) {
            $location.path("/servers");
            $route.reload();
        });
    }

    $scope.update = function () {
        $http.put('api/servers', $scope.server).success(function (data) {
            $location.path("/servers");
            $route.reload();
        });
    }
}
