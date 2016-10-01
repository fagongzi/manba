angular.module('status', []).filter("status", function(){
    return function(input){
        return input == 1 ? "UP" : "DOWN";
    };
});


/**
 *  gateway
 *
 *  Description
 */
angular.module('gateway', ["status"]).config(['$routeProvider',route]);

function route($routeProvider) {
    routeServer($routeProvider);
    routeRouting($routeProvider);
    routeAggregation($routeProvider);
    routeCluster($routeProvider);
    routeDashboard($routeProvider);
    routeProxy($routeProvider);
}
