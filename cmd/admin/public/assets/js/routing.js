function routeRouting($routeProvider) {
    $routeProvider.when("/routings", {
        "templateUrl" : "html/routing/list.html",
        "controller" : RoutingController
    }).when("/new/routing", {
        "templateUrl" : "html/routing/new.html",
        "controller" : RoutingCreateController
    });
}

function RoutingCreateController($scope, $routeParams, $http, $location, $route) {
    $http.get("api/clusters").success(function (data) {
        $scope.clusters = data.value
    });

    $scope.add = function() {
        d = {
            "clusterName": $scope.newClusterName,
            "cfg": $scope.newRoutingCfg,
            "url": $scope.newRoutingUrl
        };

        $http.post('api/routings', d).success(function(data){
            $location.path("/routings");
            $route.reload();
        });
    }
}

function RoutingController($scope, $routeParams, $http, $location, $route) {
    $http.get("api/routings").success(function (data) {
        $scope.routings = data.value
    });

    $scope.create = function() {
        $location.path("/new/routing");
    }
}
