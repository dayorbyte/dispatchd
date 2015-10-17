
REFRESH_INTERVAL = 2000

var app = angular.module('adminApp', [])

app.service('serverState', ['$http', '$interval',
function($http, $interval) {
  var self = this
  self.server = {}
  self.refresh = function() {
    $http({
      'url' : '/api/server'
    }).then(function(data) {
      angular.forEach(self.server, function(value, key) {
        delete self.server[key]
      })
      angular.forEach(data.data, function(value, key) {
        self.server[key] = value
      })
    })
  }
  self.refresh()
  $interval(self.refresh, REFRESH_INTERVAL)
}])

app.controller('ServerController', ['$scope', 'serverState',
function($scope, serverState) {
  $scope.server = serverState.server
}])
