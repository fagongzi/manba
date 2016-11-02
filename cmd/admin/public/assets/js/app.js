angular.module('status', []).filter("status", function () {
    return function (input) {
        return input == 1 ? "UP" : "DOWN";
    };
});

/**
 *  gateway
 *
 *  Description
 */
var app = angular.module('gateway', ["status"]);
app.directive('jsonText', function () {
    return {
        restrict: 'A',
        require: 'ngModel',
        link: function (scope, element, attr, ngModel) {
            function into(input) {
                return JSON.parse(input);
            }
            function out(data) {
                return JSON.stringify(data);
            }
            ngModel.$parsers.push(into);
            ngModel.$formatters.push(out);
        }
    };
});

app.config(['$routeProvider', route]);

function route($routeProvider) {
    routeServer($routeProvider);
    routeRouting($routeProvider);
    routeAPI($routeProvider);
    routeCluster($routeProvider);
    routeDashboard($routeProvider);
    routeProxy($routeProvider);
}
