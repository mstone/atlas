(function($){
$(document).ready(function(){
  var loadFragment = function(){
    // examine fragment for search terms
    var patSearchFrag = RegExp("search=([^&]*)");
    var match = patSearchFrag.exec(location.hash)
    if (match) {
      $("#searchbox").val(match[1]);
      $("#searchbox").trigger('keyup');
    }
  }
  var site;
  // $.getJSON('@APPROOT@' + '/site.json', function(data){
  $.getJSON('/site.json', function(data){
    site = data;
    $("#searchbar").css("display", "inline-block");
    loadFragment();
  });
  var makeLink;
  makeLink = function(k, v) {
    var text = v.split(/\n/)[0].replace(/^% /, '');
    //var link = "@APPROOT@" + k;
    var link = "/" + k;
    return $("<a/>", {"href": link, "text": text});
  }
  var doSearch;
  doSearch = function(){
    var search;
    search = $("#searchbox").val();
    if (search === ".") {
      var cont = $("<ul/>");
      $.each(site, function(k, v){
        cont.append($("<li/>").append(makeLink(k, v)));
      });
      $("#searchresults").empty().append(
          $("<br/><h2>Matching Charts</h2>")
        , cont
        );
    }
    else {
      if (search.length > 0) {
        var pat = RegExp(search, "i");
        var results = $("<ul>");
        $.each(site, function(k, v){
          if (v.match(pat)) {
            var pat2 = RegExp("^(.*)(" + search + ")(.*)$", "igm");
            var elt = $("<li/>").append(makeLink(k, v));
            var cont = $("<ul/>");
            var count = 0;
            var match;
            while(match = pat2.exec(v))
            {
              count++;
              if (count > 3) { break; }
              cont.append(
                $("<li/>").append("..."
                  , $("<span/>", {"class": "searchctx", "text": match[1]})
                  , $("<span/>", {"class": "searchhit", "text": match[2]})
                  , $("<span/>", {"class": "searchctx", "text": match[3]})
                  , "..."
                  )
                );
            }
            elt.append(cont);
            results.append(elt);
          }
        });
        if (results.length > 0) {
          $("#searchresults").empty().append('<br/><h2>Matching Charts</h2>', results);
        } else {
          $("#searchresults").empty().append('<br/><h2>Matching Charts</h2>', $('<p><b>None</b></p>'));
        }
      }
      else {
        $("#searchresults").empty();
      }
    }
  };
  $("#searchbox").keyup(doSearch);
  loadFragment();
});})(jQuery);
