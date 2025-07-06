# Changelog

## 1.0.2

* added distance between icons and label in tree view
* fixed text editor buttons

## 1.0.1

* fixed z-index for windows

## 1.0.0

* added admin UI
* added default page and markup when no content is present 
* always use file system provider if dev mode is enabled
* updated dependencies

## 0.13.0

* removed dir parameter from `Server.Start`

## 0.12.0

* added local and remote file synchronization
* improved configuration format
* added secrets configuration
* removed auto-updates for configuration (wasn't really useful)
* updated dependencies

## 0.11.2

* added args to custom handler

## 0.11.1

* fixed unused parameter

## 0.11.0

* allow matching routes
* allow overwriting default 404-page routes
* added an error message when folders do not match expected project structure
* added support for translated 404 error pages
* updated dependencies

## 0.10.1

* updated dependencies

## 0.10.0

* fixed concurrent data access
* updated Go version
* updated dependencies

## 0.9.2

* changed page log time from ns to ms

## 0.9.1

* fixed blocking analytics requests

## 0.9.0

* fixed configuration in readme
* upgraded to Go version 1.23
* updated dependencies

## 0.8.2

* fixed custom handler lock

## 0.8.1

* fixed creating Pirsch analytics client when client secret is not configured
* updated dependencies

## 0.8.0

* added page cache
* added log level configuration
* added logging page render time in debug mode
* updated dependencies

## 0.7.5

* fixed deadlock when updating content in dev mode
* fixed waiting for analytics provider when sending page views
* updated Go version
* updated dependencies

## 0.7.4

* fixed loading refs before page content and experiments
* fixed deadlock
* updated Go version
* updated dependencies

## 0.7.3

* fixed hostname in canonical link

## 0.7.2

* fixed server setup

## 0.7.1

* fixed sitemap setup

## 0.7.0

* exposed server sitemap

## 0.6.2

* fixed router setup

## 0.6.1

* fixed router setup

## 0.6.0

* added optional router to server options

## 0.5.0

* exported new server struct
* exposed server router and content
* fixed concurrency issue

## 0.4.0

* fixed concurrency issue

## 0.3.0

* fixed loading templates if dev is set to false
* increased fs provider reload timer

## 0.2.0

* fixed lower case paths

## 0.1.0

Initial release.
