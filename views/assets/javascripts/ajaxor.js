function initMeta(inputId, baseUrl, resourceName, metaName) {
  $("#" + inputId).select2({
    ajax: {
      url: baseUrl + "/!metas/" + resourceName + "/" + metaName,
      cache: true,
      delay: 250,
      data: function(params) {
        return {
          query: params.term,
          query_page: params.page,
        };
      },
      processResults: function(data, params) { 
        params.page = params.page || 0;
        return {
          results: data.collection ? data.collection.map(function(el) {
            return { id: el[0], text: el[1] };          
          }) : [],
          pagination: { 
            more: data.collection && data.collection.length == 20
          }
        };
      }
    }
  });
}
