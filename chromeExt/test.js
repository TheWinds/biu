　var name = "The Window";
　　var object = {
　　　　name : "My Object",
　　　　getNameFunc : function(){
　　　　　　return function(){
　　　　　　　　return this.namse;
　　　　　　};
　　　　}
　　};
console.log(object.getNameFunc()());