function routeProxy($routeProvider) {
    $routeProvider.when("/proxies", {
        "templateUrl" : "html/proxies/list.html",
        "controller" : ProxyController
    });
}

function ProxyController($scope, $routeParams, $http, $location, $route) {
    $scope.logLevel = ""

    $http.get("api/proxies").success(function (data) {
        $scope.proxies = data.value
    });

    $scope.setLogLevel = function(addr) {
        var url = "api/proxies/" + addr + "/" + $scope.logLevel

        $http.post(url, {}).success(function(data){
            $location.path("/proxies");
            $route.reload();
        });
    }
}