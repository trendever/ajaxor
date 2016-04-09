package ajaxor

import (
	"fmt"
	"github.com/qor/admin"
	"github.com/qor/qor/utils"
	"github.com/qor/roles"
	"reflect"
	"strconv"
)

var (
	// binded resources collection
	resources = map[string]*admin.Resource{}

	// size of page
	// edit carefully: it's independently set in ajax_select.js
	pageSize = 20
)

// Meta makes select_one/select_many/collection meta to use AJAX searching
func Meta(res *admin.Resource, meta *admin.Meta) {

	res.Meta(meta) // first, pass it; qor will set up GetCollection, Valuer/Setter and etc

	// check if we already registered it
	if _, registered := resources[res.Name]; !registered {
		// register && remember it
		resources[res.Name] = res
		register(res)
	}

	switch meta.Type { // now, we can change template to ours
	case "select_one":
		meta.Type = "ajaxor_select_one"
	case "select_many":
		meta.Type = "ajaxor_select_many"
	default:
		utils.ExitWithMsg(fmt.Errorf("Incorrect metas.Meta meta type: %v!", meta.Type))
	}
}

func init() {
	// register path to our templates; javascripts; stylesheets
	admin.RegisterViewPath("github.com/trendever/ajaxor/views")
}

// Init initializes ajaxor
func Init(adm *admin.Admin) {
	adm.RegisterFuncMap("resource_name", resourceName)
}

// register router handlers
func register(res *admin.Resource) {
	// load js files
	res.UseTheme("select2.min") // jquery select2 library
	res.UseTheme("ajaxor")      // our initialization code

	router := res.GetAdmin().GetRouter()
	router.Get(fmt.Sprintf("/%v/%v/!metas/:resource/:name", res.ToParam(), res.ParamIDName()), getVariantsHandler)
}

// resourceName generates resourceName; that uses our meta model
//  model should implement admin.ResourceNamer interface
func resourceName(meta *admin.Meta) string {
	// follow ptr && slice
	elemType := meta.FieldStruct.Struct.Type
	for elemType.Kind() == reflect.Slice || elemType.Kind() == reflect.Ptr {
		elemType = elemType.Elem()
	}

	// get empty struct
	value := reflect.New(elemType).Interface()

	// assume it's ResourceNamer -- get resource name
	if inter, ok := value.(admin.ResourceNamer); value != nil && ok {
		return inter.ResourceName()
	}

	utils.ExitWithMsg(fmt.Printf("Could not generate resource name for value %#v", value))
	return ""
}

// getVariantsHandler returns possible variants for custom select_one, select_many fields
// @TODO: check permissions
func getVariantsHandler(context *admin.Context) {

	// Ctx resource is what we are selecting in (for example, Order)
	// This handler is run from some specific order (for example, order{id:1})
	// Meta is our selector field. In our case -- Order.Customer
	// Resource is what we are selecting (Order.Customer is type User; so _must_ have a resource)

	var (
		// get resource
		resourceName = context.Request.URL.Query().Get(":resource")
		resource     = context.Admin.GetResource(resourceName)

		// get meta
		metaName = context.Request.URL.Query().Get(":name")
		meta     = context.Resource.GetMeta(metaName) // yes, meta is retrieved from ctxRes

		// get search keyword
		// they are intentionally named not standard: otherwise qor will use them to mess with ctx
		searchQuery   = context.Request.FormValue("query")
		searchPage, _ = strconv.Atoi(context.Request.FormValue("query_page"))
	)

	// Sanity checks
	if meta == nil {
		addError(context, fmt.Errorf("Meta %v not found", metaName))
	}

	if !meta.HasPermission(roles.Read, context.Context) {
		addError(context, fmt.Errorf("No permissions for meta %v"))
	}

	if resource == nil {
		addError(context, fmt.Errorf("Resource %v not found", resourceName))
	}

	searchHandler := resource.SearchHandler
	if searchHandler == nil {
		addError(context, fmt.Errorf("Resource %v has no search handler; did you forget to make res.SearchAttrs()?", resource.Name))
	}

	// find selected record (we work in it's context)
	record, err := context.FindOne()
	addError(context, err)

	// context we will search entries in
	searchCtx := context.Clone()
	searchCtx.SetDB(searchHandler(searchQuery, searchCtx).
		Limit(pageSize).
		Offset(searchPage * pageSize),
	)

	// do the search using meta.GetCollection
	out := meta.GetCollection(record, searchCtx)
	context.JSON("show", map[string]interface{}{"collection": out})
}

func addError(ctx *admin.Context, err error) {
	ctx.AddError(err) //@TODO: smth wrong with ret error

	if ctx.HasError() {
		ctx.JSON("show", map[string]interface{}{"errors": ctx.GetErrors()})
		utils.ExitWithMsg(err.Error())
	}
}
