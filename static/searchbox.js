(function($){
$(document).ready(function(){
  // Function for implementing bookmarkable searches.
  var loadFragment;
  loadFragment = function(){
    // examine fragment for search terms
    var patSearchFrag = RegExp("search=([^&]*)");
    var match = patSearchFrag.exec(location.hash)
    if (match) {
      $("#searchbox").val(match[1]);
      $("#searchbox").trigger('keyup');
    }
  }

  // Ref for a chart-name -> search-view map.
  var site;

  // Ref for matching (chart-names), for form submission-handling.
  // Updated by doSearch(); read by doSubmit().
  var matchingChartnames = [];

  // Use XHR to attempt to fill the site-ref.
  // $.getJSON('@APPROOT@' + '/site.json', function(data){
  $.getJSON('/site.json', function(data){
    site = data;
    $("#searchbox").attr("disabled", false);
    $("#searchbar").css("display", "inline-block");
    loadFragment();
  });

  // Helper function to convert (chart-name, search-view) to (href,
  // text)
  var makeLinkData;
  makeLinkData = function(k, v) {
    var text = v.split(/\n/)[0].replace(/^% /, '');
    //var link = "@APPROOT@" + k;
    var link = "/" + k;
    return {"href": link, "text": text};
  };

  // Return an anchor element for (chart-name, search-view)
  var makeLink;
  makeLink = function(k, v) {
    return $("<a/>", makeLinkData(k, v));
  }

  // If only one result is found, jump to it; called on
  // #searchform.submit()
  var doSubmit;
  doSubmit = function(){
    if (matchingChartNames.length > 0) {
      location.href = "/" + matchingChartNames[0];
    }
    return false;
  };

  // Calculate search results; called on #searchinput.keyup()
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
        matchingChartNames = [];
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
            matchingChartNames.push(k);
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
    $('html, body').scrollTop(0);
  };
  $("#searchform").submit(doSubmit);
  $("#searchbox").keyup(doSearch);
  loadFragment();


  // Display ticket information!
  $("h1, h2, h3, h4, h5, h6").wrap("<div style=\"clear: both;\"></div>");
  var ticketUriPrefix = "data:tkt,";
  var ticketAttrSelector = "a[href^=\"" + ticketUriPrefix + "\"]";
  var ticketSelector = ["h1 > " + ticketAttrSelector,
                        "h2 > " + ticketAttrSelector,
                        "h3 > " + ticketAttrSelector,
                        "h4 > " + ticketAttrSelector,
                        "h5 > " + ticketAttrSelector,
                        "h6 > " + ticketAttrSelector].join(", ");
  var renderTicket = function(k,v){
    var href = v.href;
    if (href.indexOf(ticketUriPrefix) == 0) {
      var qs = href.substr(ticketUriPrefix.length);
      var obj = {};
      var match;
      var kvPat = RegExp("([^=&]+)=([^&]*)&?", "g");
      while(match = kvPat.exec(qs)) {
        obj[decodeURIComponent(match[1])] = decodeURIComponent(match[2]);
      }

      $(v).attr({"display": "none"});

      var tbl = $("<table>")
                  .attr({"class": "ticket"})
                  .append($("<caption>Ticket State</caption>"));
      //tbl.append("<tr><th>Key</th><th>Value</th></tr>");
      $.each(obj, function(k,v){
          var tr = $("<tr>").append($("<td class=\"ticket-key\">").append(k))
                            .append($("<td class=\"ticket-val\">").append(v));
          tbl.append(tr);
      });
      $(v).parent().parent().append(tbl);
    }
  };
  $(ticketSelector).each(renderTicket);

});})(jQuery);
