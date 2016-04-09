# Ajaxor

## What it is 

Ajaxor is a plugin for [qor](http://getqor.com/) uses [jQuery select2](https://select2.github.io/) selectors to load only necessary entries using REST api. 

Usual `select_one` and `select_many` qor meta types are very plain: they generate `<select>` with an `<option>` for every selectable item.
While they are extremely plain and simple, they bring us very big and slow pages (2MB of HTML and 2 seconds of download for an selector with 9000 entries).
That makes them unsuitable for real enterprise use.

## How to use it

Suppose, we have the following models:

```go
type Parent struct {
  gorm.Model
  
  ChildID uint
  Child   Child
}

type Child struct {
  gorm.Model

  Name string
}
```

In order to make Ajaxor work we have to:

1) Register both resources for Parent and Child in qor.

  * Remember, that auto-generated .Meta.Resource is not registered in qor. Create own one
    * Resources can be invisible :)

2) Child model _must_ implement qor.admin.ResourceNamer interface. 

  ```go
  func (c Child) ResourceName() string {
    return "ChildResource"
  }
  ```

3) When adding the resource `res` to qor, call resource.Meta() wrapper:

  ```go
	ajaxor.Meta(res, &admin.Meta{
		Name: "Child",
		Type: "select_one",
	})
  ```

4) Ajax selects should now replace default ones

## Current issues 

* Uses our internal debug pkg
* Does not use standard material-design theme. 
* I don't like binding the routes to every resource. Maybe I should context manually?

## License

Released under the [MIT License](http://opensource.org/licenses/MIT).
