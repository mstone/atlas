(function($){
$(document).ready(function(){
  // Function for implementing bookmarkable searches.
  var loadFragment;
  loadFragment = function(){
    // examine fragment for search terms
    var patFindFrag = RegExp("find=([^&]*)");
    var matchFind = patFindFrag.exec(location.hash);

    var patSearchFrag = RegExp("search=([^&]*)");
    var matchSearch = patSearchFrag.exec(location.hash);
    if (matchFind || matchSearch) {
      $("#searchfind").val(matchFind[1] || "");
      $("#searchgrep").val(matchGrep[1] || "");
      $("#searchgrep").trigger('keyup');
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
    $("#searchfind").attr("disabled", false);
    $("#searchgrep").attr("disabled", false);
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

  var updateResults;
  updateResults = function(prefix, results) {
        $("#searchresults").empty();
        if (prefix !== null) {
          $("#searchresults").append(prefix);
        }
        if (results.length > 0) {
          $("#searchresults").append('<h2>Matching Charts</h2>', results);
        } else {
          $("#searchresults").append('<h2>Matching Charts</h2>', $('<p><b>None</b></p>'));
        }
  }

  // Calculate search results; called on #searchinput.keyup()
  var doSearch;
  doSearch = function(ev){
    // check for form submission
    if (ev.which == 13) {
      doSubmit();
      return;
    }
    var findStr; // chart url filter string
    var grepStr; // chart body filter string
    var findPat = null; // regex of chart url filter
    var grepPat; // regex of chart body filter
    findStr = $("#searchfind").val() || "";
    grepStr = $("#searchgrep").val() || "";
    matchingChartNames = [];
    if (findStr.length > 0) {
      findPat = RegExp(findStr, "i");
    }
    // if asked to list names...
    if (grepStr === "." || (findStr.length > 0 && grepStr.length == 0)) {
      var cont = $("<ul/>");
      $.each(site, function(k, v){
        if (findPat === null || k.match(findPat)) {
          matchingChartNames.push(k)
          cont.append($("<li/>").append(makeLink(k, v)));
        }
      });
      updateResults(null, cont);
    } else { // examine bodies
      if (grepStr.length > 0) {
        grepPat = RegExp(grepStr, "i");
        var results = $("<ul>");
        $.each(site, function(k, v){
          if (findPat === null || k.match(findPat)) {
            if (v.match(grepPat)) {
              matchingChartNames.push(k);
              var pat2 = RegExp("^(.*)(" + grepStr + ")(.*)$", "igm");
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
          }
        });
        var newChartLink = $("<a>make a new chart</a>").attr({href: "/" + $.trim(grepStr) + "/index.txt/editor"});
        var resultsPrefix = $('<p class="newChartLink">').append(['(Alternately, shall we ', newChartLink, ' for that?)']);
        updateResults(resultsPrefix, results);
      }
      else {
        $("#searchresults").empty();
      }
    }
    $('html, body').scrollTop(0);
  };
  $("#searchform").submit(doSubmit);
  $("#searchfind").keyup(doSearch);
  $("#searchgrep").keyup(doSearch);
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
