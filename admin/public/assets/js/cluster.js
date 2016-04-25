function routeCluster($routeProvider) {
    $routeProvider.when("/clusters", {
        "templateUrl" : "html/cluster/list.html",
        "controller" : ClusterController
    }).when("/clusters/:name", {
        "templateUrl" : "html/cluster/detail.html",
        "controller" : ClusterDetailController
    }).when("/new/cluster", {
        "templateUrl" : "html/cluster/new.html",
        "controller" : ClusterCreateController
    });
}

function ClusterDetailController($scope, $routeParams, $http, $location, $route) {
    $http.get("api/clusters/" +  $routeParams.name).success(function (data) {
        $scope.cluster = data.value;
    });

    $http.get("api/lbs").success(function (data) {
        $scope.lbs = data
    });

    $scope.update = function() {
        $http.put('api/clusters', $scope.cluster).success(function(data){
            $location.path("/clusters");
            $route.reload();
        });
    }
}

function ClusterCreateController($scope, $routeParams, $http, $location, $route) {
    $http.get("api/lbs").success(function (data) {
        $scope.lbs = data
    });

    $scope.add = function() {
        var c = {
            "name": $scope.newClusterName,
            "pattern": $scope.newClusterPattern,
            "lbName": $scope.newClusterLBName,
            "count": 0
        };

        $http.post('api/clusters', c).success(function(data){
            $location.path("/clusters");
            $route.reload();
        });
    }
}

function ClusterController($scope, $routeParams, $http, $location, $route) {
    $http.get("api/clusters").success(function (data) {
        $scope.clusters = data.value

        var len = $scope.clusters.length;
        for (var i = 0; i < len; i++) {
            $scope.clusters[i].checked = false;
        }
    });


    $scope.create = function() {
        $location.path("/new/cluster");
        $route.reload();
    }

    $scope.delete = function(name) {
        $http.delete('api/clusters/' + name).success(function(data){
            $location.path("/clusters");
            $route.reload();
        });
    }
}

