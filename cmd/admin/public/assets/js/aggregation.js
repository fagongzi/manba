function routeAggregation($routeProvider) {
    $routeProvider.when("/aggregations", {
        "templateUrl": "html/aggregation/list.html",
        "controller": AggregationController
    }).when("/aggregations/:url", {
        "templateUrl" : "html/aggregation/update.html",
        "controller" : AggregationUpdateController
    }).when("/new/aggregation", {
        "templateUrl": "html/aggregation/new.html",
        "controller": AggregationCreateController
    });
}

function AggregationUpdateController($scope, $routeParams, $http, $location, $route) {
    $http.get("api/clusters").success(function (data) {
        $scope.clusters = data.value;
    });

    $http.get("api/aggregation?url=" + encodeURIComponent(atob($routeParams.url))).success(function (data) {
        $scope.ang = data.value;
    });

    $scope.newUrl = "";


    $scope.resetNode = function () {
        $scope.newClusterName = "";
        $scope.newNodeUrl = "";
        $scope.newNodeAttrName = "";
        $scope.newNodeRewrite = "";
    }

    $scope.addNode = function () {
        $scope.ang.nodes.push({
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
        $http.put('api/aggregations', $scope.ang).success(function (data) {
            $location.path("/aggregations");
            $route.reload();
        });
    }
}


function AggregationCreateController($scope, $routeParams, $http, $location, $route) {
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

        $http.post('api/aggregations', d).success(function (data) {
            $location.path("/aggregations");
            $route.reload();
        });
    }
}

function AggregationController($scope, $routeParams, $http, $location, $route) {
    $http.get("api/aggregations").success(function (data) {
        $scope.aggregations = data.value;
        
        for(var i = 0; i < $scope.aggregations.length; i++) {
            $scope.aggregations[i].encodeURL = btoa($scope.aggregations[i].url);
        }
    });

    $scope.create = function () {
        $location.path("/new/aggregation");
    }

    $scope.delete = function (url) {
        $http.delete('api/aggregations?url=' + encodeURIComponent(url)).success(function (data) {
            $location.path("/aggregations");
            $route.reload();
        });
    }
}