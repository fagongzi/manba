function routeAPI($routeProvider) {
    $routeProvider.when("/apis", {
        "templateUrl": "html/api/list.html",
        "controller": APIController
    }).when("/apis/:url", {
        "templateUrl" : "html/api/update.html",
        "controller" : APIUpdateController
    }).when("/new/api", {
        "templateUrl": "html/api/new.html",
        "controller": APICreateController
    });
}

function APIUpdateController($scope, $routeParams, $http, $location, $route) {
    $http.get("api/clusters").success(function (data) {
        $scope.clusters = data.value;
    });

    $http.get("api/apis/" + $routeParams.url).success(function (data) {
        $scope.api = data.value;
    });

    $scope.newUrl = "";


    $scope.resetNode = function () {
        $scope.newClusterName = "";
        $scope.newNodeUrl = "";
        $scope.newNodeAttrName = "";
        $scope.newNodeRewrite = "";
    }

    $scope.addNode = function () {
        $scope.api.nodes.push({
            "clusterName": $scope.newClusterName,
            "url": $scope.newNodeUrl,
            "attrName": $scope.newNodeAttrName,
            "rewrite": $scope.newNodeRewrite
        });

        $scope.resetNode();
    }

    $scope.delete = function (node) {
        ns = []

        for (var i = 0; i < $scope.ang.nodes.length; i++) {
            if ($scope.ang.nodes[i] != node) {
                ns.push($scope.ang.nodes[i]);
            }
        }

        $scope.ang.nodes = ns;
    }

    $scope.update = function () {
        $http.put('api/apis', $scope.ang).success(function (data) {
            $location.path("/apis");
            $route.reload();
        });
    }
}


function APICreateController($scope, $routeParams, $http, $location, $route) {
    $http.get("api/clusters").success(function (data) {
        $scope.clusters = data.value;
    });

    $scope.newUrl = "";
    $scope.newNodes = [];


    $scope.resetNode = function () {
        $scope.newClusterName = "";
        $scope.newNodeUrl = "";
        $scope.newNodeAttrName = "";
        $scope.newNodeRewrite = "";
    }

    $scope.addNode = function () {
        $scope.newNodes.push({
            "clusterName": $scope.newClusterName,
            "url": $scope.newNodeUrl,
            "attrName": $scope.newNodeAttrName,
            "rewrite": $scope.newNodeRewrite
        });

        $scope.resetNode();
    }

    $scope.delete = function (node) {
        ns = []

        for (var i = 0; i < $scope.newNodes.length; i++) {
            if ($scope.newNodes[i] != node) {
                ns.push($scope.newNodes[i]);
            }
        }

        $scope.newNodes = ns;
    }

    $scope.add = function () {
        d = {
            "url": $scope.newUrl,
            "nodes": $scope.newNodes,
        }

        $http.post('api/apis', d).success(function (data) {
            $location.path("/apis");
            $route.reload();
        });
    }
}

function APIController($scope, $routeParams, $http, $location, $route) {
    $http.get("api/apis").success(function (data) {
        $scope.apis = data.value;
        
        for(var i = 0; i < $scope.apis.length; i++) {
            $scope.apis[i].encodeURL = btoa($scope.apis[i].url);
        }
    });

    $scope.create = function () {
        $location.path("/new/api");
    }

    $scope.delete = function (url) {
        $http.delete('api/apis/' + url).success(function (data) {
            $location.path("/apis");
            $route.reload();
        });
    }
}