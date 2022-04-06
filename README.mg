# Setup local dtrack
* For a fresh clone please run bootstrap.
```
make bootstrap-dtrack
```

* Start local (First run may take a while)
```
make start-local
```

* Enter frontend in browser
```
http://localhost:8080/
```

* First login admin:admin (change password)

* Enter `projects` in side bar -> create project.

* Select project from project list -> `Components` tab

* Select `Upload BOM` -> select SBOM to upload.

Note: BOM analysis (components and then graph) may takes a while for large sboms.

