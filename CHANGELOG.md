# Changelog
All notable changes to this project will be documented in this file.

## 3.19.0 - 2026-04-22
### Features
* Added support for variation simulation when feature flags are **inactive (OFF state)** across the following methods:
    - [`GetVariation`](https://developers.kameleoon.com/feature-management-and-experimentation/web-sdks/go-sdk/#getvariation)
    - [`GetVariations`](https://developers.kameleoon.com/feature-management-and-experimentation/web-sdks/go-sdk/#getvariations)
    - [`IsFeatureActive`](https://developers.kameleoon.com/feature-management-and-experimentation/web-sdks/go-sdk/#isfeatureactive--isfeatureactivewithtracking)
    - This support is also extended to the following **deprecated methods**:
        - [`GetActiveFeatureListForVisitor`](https://developers.kameleoon.com/feature-management-and-experimentation/web-sdks/go-sdk/#getactivefeaturelistforvisitor)
        - [`GetActiveFeatures`](https://developers.kameleoon.com/feature-management-and-experimentation/web-sdks/go-sdk/#getactivefeatures)
        - [`GetFeatureVariationKey`](https://developers.kameleoon.com/feature-management-and-experimentation/web-sdks/go-sdk/#getfeaturevariationkey)
        - [`GetFeatureVariable`](https://developers.kameleoon.com/feature-management-and-experimentation/web-sdks/go-sdk/#getfeaturevariable)
* Improved the performance of the [`getDataFile()`](https://developers.kameleoon.com/feature-management-and-experimentation/web-sdks/go-sdk/#getdatafile) instance that is refreshed whenever the SDK configuration is updated.
* The [`getFeatureList`](https://developers.kameleoon.com/feature-management-and-experimentation/web-sdks/go-sdk/#getfeaturelist) method has been deprecated in favor of `getDataFile()`.

## 3.18.0 - 2026-02-13
### Features
* Updated the allowed range for the [`TrackingInterval`](https://developers.kameleoon.com/feature-management-and-experimentation/web-sdks/go-sdk/#additional-configuration). The new range is from **`1000` ms** (default) to **`5000` ms**, allowing a reduction in the number of tracking requests.
* Introduced a new `track` parameter for [`AddData`](https://developers.kameleoon.com/feature-management-and-experimentation/web-sdks/go-sdk/#adddata). When set to `false`, the data is stored locally and used only for targeting evaluation; it is not sent to the Data API, helping to prevent duplicate data from being recorded. The default value is `true`. This behavior is consistent with the `track` parameter used in evaluation methods such as [`GetVariation`](https://developers.kameleoon.com/feature-management-and-experimentation/web-sdks/go-sdk/#getvariation).
* Introduced support for the Kameleoon [`ApplicationVersion`](https://developers.kameleoon.com/feature-management-and-experimentation/web-sdks/go-sdk/#applicationversion) data type, allowing the visitor’s application version to be set and used in targeting evaluation.

## 3.17.0 - 2025-12-24
### Features
* Introduced a new [`GetDataFile`](https://developers.kameleoon.com/feature-management-and-experimentation/web-sdks/go-sdk#getdatafile) method. This method returns the current SDK configuration (also known as the **data file**) used for evaluation and targeting. It is **not** intended for production use to fetch variations for every feature flag in the returned list, as it is not optimized for performance. For that purpose, use [`GetVariations`](https://developers.kameleoon.com/feature-management-and-experimentation/web-sdks/go-sdk#getvariations) instead. `GetDataFile` is mainly useful for debugging or QA, for example to let internal users manually select a variant for a specific feature flag in production.
### Bug fixes
* Fixed an issue where **legal consent** was not taken into account during holdout evaluation.

## 3.16.0 - 2025-11-21
### Features
* Updated evaluation and tracking logic to comply with GDPR requirements when consent is not given:
    - If behavior is **partially blocked**, the default variation will be returned.
    - If behavior is **completely blocked**, an exception will be thrown.

## 3.15.1 - 2025-10-31
### Bug fixes
* Fixed an issue where [`Conversion`](https://developers.kameleoon.com/feature-management-and-experimentation/web-sdks/go-sdk#conversion)'s metadata initialized with a name was not tracked.

## 3.15.0 - 2025-08-29
### Features
* Added an `overwrite` flag to [`CustomData`](https://developers.kameleoon.com/feature-management-and-experimentation/web-sdks/go-sdk#customdata), used as the `overwrite` parameter during tracking.
* [`CustomData`](https://developers.kameleoon.com/feature-management-and-experimentation/web-sdks/go-sdk#customdata) can now be created using a `name`, in addition to the existing method of using an `index`.

## 3.14.0 - 2025-07-23
### Features
* Added the [`EvaluateAudiences`](https://developers.kameleoon.com/feature-management-and-experimentation/web-sdks/go-sdk#evaluateaudiences) method. This method iterates over all Audiences Explorer segments, evaluates each one, and tracks the segments for which the visitor is targeted using the [`TARGETINGSEGMENT`](https://developers.kameleoon.com/apis/data-api-rest/all-endpoints/post-visit-events/) event.

## 3.13.0 - 2025-06-27
### Features
* Added support for using a Custom Data value for traffic bucketing instead of the default visitor code. [Learn More](https://help.kameleoon.com/create-feature-flag/#Advanced_Flag_Settings).

## 3.12.0 - 2025-05-26
### Features
* Added support for **304 (Not Modified)** responses from the SDK config service to avoid redundant updates and reduce traffic when the configuration hasn't changed.
* Added support for a **New**/**Returning** visitor breakdown filter in reports (requires calling [`GetRemoteVisitorData`](https://developers.kameleoon.com/feature-management-and-experimentation/web-sdks/go-sdk#getremotevisitordata)).
### Fixed
* Fixed an issue where visitor data fields - [`Browser`](https://developers.kameleoon.com/feature-management-and-experimentation/web-sdks/go-sdk#browser), [`Device`](https://developers.kameleoon.com/feature-management-and-experimentation/web-sdks/go-sdk#device), and [`OperatingSystem`](https://developers.kameleoon.com/feature-management-and-experimentation/web-sdks/go-sdk#operatingsystem) - were all retrieved from the Data API and added to the visitor, even when only a subset of them was requested.

## 3.11.1 - 2025-04-08
### Bug fixes
* Changed the order in which **conversion** and **experiment** events are sent. This may lead to more accurate **visit**-level experiment reporting.

## 3.11.0 - 2025-03-24
### Features
* Added new variation of the [`TrackConversion`](https://developers.kameleoon.com/feature-management-and-experimentation/web-sdks/go-sdk#trackconversion) method:
  - `TrackConversionWithOptParams(visitorCode string, goalId int, params TrackConversionOptParams) error`
* Added new variation of the [`Conversion`](https://developers.kameleoon.com/feature-management-and-experimentation/web-sdks/go-sdk#conversion) data constructor:
  - `NewConversionWithOptParams(goalId int, params ConversionOptParams) *Conversion`

## 3.10.0 - 2025-03-18
### Features
* Added support for Contextual Bandit evaluations. Calling [`GetRemoteVisitorData`](https://developers.kameleoon.com/feature-management-and-experimentation/web-sdks/go-sdk#getremotevisitordata) with the `cbs=true` flag is required for this feature to function correctly. Platform-wide release expected in March 2025.
* Added new configuration parameter `NetworkDomain` (`network_domain`) to [`KameleoonClientConfig`](https://developers.kameleoon.com/feature-management-and-experimentation/web-sdks/go-sdk/#create) and the [configuration](https://developers.kameleoon.com/go-sdk.html#additional-configuration) file. This parameter allows specifying a custom domain for all outgoing network requests.
* Added support for new conditions:
    - Exclusive Campaign
    - Experiment
    - Personalization

## 3.9.0 - 2025-02-26
### Features
* Added SDK support for **Mutually Exclusive Groups**. When feature flags are grouped into a **Mutually Exclusive Group**, only one flag in the group will be evaluated at a time. All other flags in the group will automatically return their default variation.

## 3.8.0 - 2025-02-10
### Features
* Added SDK support for **holdout experiments**. Visitors assigned to a holdout experiment are excluded from all other rollouts and experiments, and consistently receive the default variation. For visitors not in a holdout experiment, the standard evaluation process applies, allowing them to be evaluated for all feature flags as usual. Platform-wide release expected in February 2025.
### Bug fixes
* Fixed an issue in the [`GetActiveFeatureListForVisitor`](https://developers.kameleoon.com/feature-management-and-experimentation/web-sdks/go-sdk#getactivefeaturelistforvisitor) method where feature flags disabled for the environment were not being filtered out.

## 3.7.0 - 2024-12-16
### Features
* Added support for **simulated** variations.
* Added the [`SetForcedVariation()`](https://developers.kameleoon.com/feature-management-and-experimentation/web-sdks/go-sdk#setforcedvariation) method. This method allows explicitly setting a forced variation for a visitor, which will be applied during experiment evaluation.
### Bug fixes
* Fixed the [`Variation.IsActive()`](https://developers.kameleoon.com/feature-management-and-experimentation/web-sdks/go-sdk#variation) method, which was returning an incorrect value.
* Fixed an issue where the return values in the trace logs for the following methods of the `kameleoonClient` were always default:
  - `IsFeatureActive`
  - `IsFeatureActiveWithTracking`
  - `isFeatureActive`
  - `GetVariation`
  - `GetVariations`
  - `makeExternalVariation`

## 3.6.1 - 2024-11-20
### Bug fixes
* Resolved an issue where the validation of [top-level domains](https://developers.kameleoon.com/feature-management-and-experimentation/web-sdks/go-sdk/#additional-configuration) for `localhost` resulted in incorrect failures. The SDK now accepts the provided domain without modification if it is deemed invalid and logs an [error](https://developers.kameleoon.com/feature-management-and-experimentation/web-sdks/go-sdk/#log-levels) to notify you of any issues with the specified domain.

## 3.6.0 - 2024-11-14
### Features
* Introduced a new `VisitorCode` parameter to [`RemoteVisitorDataFilter`](https://developers.kameleoon.com/feature-management-and-experimentation/web-sdks/go-sdk/#using-parameters-in-getremotevisitordata). This parameter determines whether to use the `VisitorCode` from the most recent previous visit instead of the current `VisitorCode`. When enabled, this feature allows visitor exposure to be based on the retrieved `VisitorCode`, facilitating [cross-device reconciliation](https://developers.kameleoon.com/core-concepts/cross-device-experimentation/). Default value of the parameter is `true`.
### Bug fixes
* Fixed an issue with the [`Page URL`](https://developers.kameleoon.com/feature-management-and-experimentation/using-visit-history-in-feature-flags-and-experiments/#benefits-of-calling-getremotevisitordata) and [`Page Title`](https://developers.kameleoon.com/feature-management-and-experimentation/using-visit-history-in-feature-flags-and-experiments/#benefits-of-calling-getremotevisitordata) targeting conditions, where the condition evaluated all previously visited URLs in the session instead of only the current URL, corresponding to the latest added [`PageView`](https://developers.kameleoon.com/feature-management-and-experimentation/web-sdks/go-sdk/#pageview).
**NOTE**: This change may impact your existing targeting. Please review your targeting conditions to ensure accuracy.

## 3.5.0 - 2024-10-04
### Features
* Introduced new evaluation methods for clarity and improved efficiency when working with the SDK:
  - [`GetVariation`](https://developers.kameleoon.com/feature-management-and-experimentation/web-sdks/go-sdk/#getvariation)
  - [`GetVariations`](https://developers.kameleoon.com/feature-management-and-experimentation/web-sdks/go-sdk/#getvariations)
* These methods replace the deprecated ones:
  - [`GetFeatureVariationKey`](https://developers.kameleoon.com/feature-management-and-experimentation/web-sdks/go-sdk#getfeaturevariationkey)
  - [`GetFeatureVariable`](https://developers.kameleoon.com/feature-management-and-experimentation/web-sdks/go-sdk#getfeaturevariable)
  - [`GetActiveFeatures`](https://developers.kameleoon.com/feature-management-and-experimentation/web-sdks/go-sdk#getactivefeatures)
  - [`GetFeatureVariationVariables`](https://developers.kameleoon.com/feature-management-and-experimentation/web-sdks/go-sdk#getfeaturevariationvariables)
* A new method [`IsFeatureActiveWithTracking`](https://developers.kameleoon.com/feature-management-and-experimentation/web-sdks/go-sdk/#isfeatureactive--isfeatureactivewithtracking) includes the `track` parameter, which controls whether the assigned variation is tracked.
* Enhanced top-level domain validation within the SDK. The implementation now includes automatic trimming of extraneous symbols and provides a [warning](https://developers.kameleoon.com/feature-management-and-experimentation/web-sdks/go-sdk/#log-levels) when an invalid domain is detected.
* Enhanced the [`GetEngineTrackingCode`](https://developers.kameleoon.com/feature-management-and-experimentation/web-sdks/go-sdk#getenginetrackingcode) method to properly handle `JS` and `CSS` variables.
### Bug fixes
* Fixed returning the error when creating a new client was unsuccessful

## 3.4.0 - 2024-08-15
### Features
* Improved the tracking mechanism to consolidate multiple visitors into a single request. The new approach combines information on all affected visitors into one request, which is sent once per interval.
  - Flush changes:
    - The [`FlushVisitor`](https://developers.kameleoon.com/feature-management-and-experimentation/web-sdks/go-sdk/#flushall--flushvisitor--flushvisitorinstantly) now enqueues the visitor's data to be tracked with next tracking interval.
    - Added a new [`FlushVisitorInstantly`](https://developers.kameleoon.com/feature-management-and-experimentation/web-sdks/go-sdk/#flushall--flushvisitor--flushvisitorinstantly) method which tracks the visitor's data instantly.
    - Added a new parameter `instant` of the [`FlushAll`](https://developers.kameleoon.com/feature-management-and-experimentation/web-sdks/go-sdk/#flushall--flushvisitor--flushvisitorinstantly) method. If the parameter's value is `true` the visitor's data is tracked instantly. Otherwise, the visitor's data will be tracked with next tracking interval. Default value of the parameter is `false`.
* Added new configuration parameter `TrackingInterval` (`tracking_interval`) to [`KameleoonClientConfig`](https://developers.kameleoon.com/go-sdk.html#initializing-the-kameleoon-client) and the [configuration](https://developers.kameleoon.com/go-sdk.html#additional-configuration) file, which is used to set interval for tracking requests. Default value is `1000` milliseconds.
* New Kameleoon Data type [`UniqueIdentifier`](https://developers.kameleoon.com/feature-management-and-experimentation/web-sdks/csharp-sdk#uniqueidentifier) is introduced. It will be used in all methods instead of `isUniqueIdentifier` parameter.
  - The `isUniqueIdentifier` parameter is marked as deprecated for the following methods:
    - [`FlushVisitor`](https://developers.kameleoon.com/feature-management-and-experimentation/web-sdks/go-sdk/#flushall--flushvisitor--flushvisitorinstantly)
    - [`TrackConversion`](https://developers.kameleoon.com/feature-management-and-experimentation/web-sdks/go-sdk/#trackconversion)
    - [`GetFeatureVariationKey`](https://developers.kameleoon.com/feature-management-and-experimentation/web-sdks/go-sdk#getfeaturevariationkey)
    - [`GetFeatureVariable`](https://developers.kameleoon.com/feature-management-and-experimentation/web-sdks/go-sdk#getfeaturevariable)
    - [`IsFeatureActive`](https://developers.kameleoon.com/feature-management-and-experimentation/web-sdks/go-sdk#isfeatureactive)
  - The [`GetRemoteVisitorDataWithOptParams`](https://developers.kameleoon.com/feature-management-and-experimentation/web-sdks/go-sdk#getremotevisitordata) method is deprecated. Please use the [`GetRemoteVisitorDataWithFilter`](https://developers.kameleoon.com/feature-management-and-experimentation/web-sdks/go-sdk#getremotevisitordata) method instead.
### Bug fixes
* The SDK no longer logs failed tracking requests to the [Data API](https://developers.kameleoon.com/apis/data-api-rest/all-endpoints/post-visit-events/) when the user agent is identified as a bot (i.e., when the status code is 403).
* Fixed an issue that caused duplicate entries in feature flag results for both anonymous and authorized/identified visitors during data reconciliation. This problem occurred when custom data of type mapping ID was not consistently sent for all sessions.

## 3.3.0 - 2024-06-21
### Features
* Added [`GetActiveFeatures`](https://developers.kameleoon.com/go-sdk.html#getactivefeatures) method. It retrieves information about the active feature flags that are available for a specific visitor code. This method replaces the deprecated [`GetActiveFeatureListForVisitor`](https://developers.kameleoon.com/go-sdk.html#getactivefeaturelistforvisitor) method.
### Bug fixes
* The SDK no longer logs failed tracking requests to the [Data API](https://developers.kameleoon.com/apis/data-api-rest/all-endpoints/post-visit-events/) when the user agent is identified as a bot (i.e., when the status code is 403).

## 3.2.0 - 2024-06-10
### Features
* New targeting conditions are now available (some of them may require [`GetRemoteVisitorData`](https://developers.kameleoon.com/feature-management-and-experimentation/web-sdks/go-sdk#getremotevisitordata) pre-loaded data)
  - Browser Cookie
  - Operating System
  - IP Geolocation
  - Kameleoon Segment
  - Target Feature Flag
  - Previous Page
  - Number of Page Views
  - Time since First Visit
  - Time since Last Visit
  - Number of Visits Today
  - Total Number of Visits
  - New or Returning Visitor
  - Likelihood to convert
* New Kameleoon Data types were introduced:
  - [`Cookie`](https://developers.kameleoon.com/feature-management-and-experimentation/web-sdks/go-sdk#cookie)
  - [`OperatingSystem`](https://developers.kameleoon.com/feature-management-and-experimentation/web-sdks/go-sdk#operatingsystem)
  - [`Geolocation`](https://developers.kameleoon.com/feature-management-and-experimentation/web-sdks/go-sdk#geolocation)
### Bug fixes
* Stability and performance improvements

## 3.1.0 - 2024-02-29
### Features
* Added support for additional Data API servers across the world for even faster network requests.
* Increased limit for requests to Data API: [rate limits](https://developers.kameleoon.com/apis/data-api-rest/overview/#rate-limits)
* Added [`GetVisitorWarehouseAudience`](https://developers.kameleoon.com/go-sdk.html#getvisitorwarehouseaudience) method to retrieve all data associated with a visitor's warehouse audiences and adds it to the visitor.

## 3.0.3 - 2023-12-06
### Bug fixes
* Stability and performance improvements

## 3.0.2 - 2023-11-27
### Bug fixes
* Stability and performance improvements

## 3.0.1 - 2023-11-24
### Bug fixes
* Stability and performance improvements

## 3.0.0 - 2023-11-24
### Breaking changes
* Increased the minimum required version of the Go language to [1.18](https://go.dev/doc/go1.18).
* Renamed `Client` to `KameleoonClient`
* Removed `NewClient` function. Instead, use [`KameleoonClientFactory`](https://developers.kameleoon.com/feature-management-and-experimentation/web-sdks/go-sdk/#kameleoonclientfactory)
* Removed `RunWhenReady` method. Instead, use [`WaitInit`](https://developers.kameleoon.com/feature-management-and-experimentation/web-sdks/go-sdk/#waitinit)
* Changed `Config`:
    - Renamed `Config` to `KameleoonClientConfig`
    - Changed the default request timeout to `10` seconds
    - Renamed the `Timeout` and `timeout` configuration fields to `DefaultTimeout` and [`default_timeout`](https://developers.kameleoon.com/feature-management-and-experimentation/web-sdks/go-sdk#additional-configuration), respectively)
    - Renamed the `ConfigUpdateInterval` and `config_update_interval` configuration fields to `RefreshInterval` and [`refresh_interval`](https://developers.kameleoon.com/feature-management-and-experimentation/web-sdks/go-sdk#additional-configuration), respectively.)
    - `LoadConfig` function and `KameleoonClientConfig.Load` can now return an error.
* The `Cfg` field is no longer accessible.
* Renamed `GetFeatureAllVariables` method to `GetFeatureVariationVariables`(https://developers.kameleoon.com/feature-management-and-experimentation/web-sdks/go-sdk/#getfeaturevariationvariables)
* Removed all methods and errors related to **experiments**:
  * Methods:
    - `TriggerExperiment`
    - `GetVariationAssociatedData`
    - `GetExperimentList`
    - `GetExperimentListForVisitor`
  * Error types:
    - `ErrExperimentConfigNotFound`
    - `ErrNotTargeted`
    - `ErrNotAllocated`
    - `ErrSiteCodeDisabled`
* Changed errors:
  * Moved error types into `errs` package
  * Added `UnexpectedStatusCode` error, which can be thrown by the following methods:
    - [`GetRemoteData`](https://developers.kameleoon.com/feature-management-and-experimentation/web-sdks/go-sdk/#getremotedata)
    - [`GetRemoteVisitorData`](https://developers.kameleoon.com/feature-management-and-experimentation/web-sdks/go-sdk/#getremotevisitordata)
  * Renamed the following errors:
    - `ErrFeatureConfigNotFound` to `FeatureNotFound`
    - `ErrFeatureVariableNotFound` to `FeatureVariableNotFound`
    - `ErrVariationNotFound` to `FeatureVariationNotFound`
    - `ErrCredentialsNotFound` to `ConfigCredentialsInvalid`
    - `ErrVisitorCodeNotValid` to `VisitorCodeInvalid`
* The new error `FeatureEnvironmentDisabled` may be returned by the following methods:
    - [`GetFeatureVariationKey`](https://developers.kameleoon.com/feature-management-and-experimentation/web-sdks/go-sdk/#getfeaturevariationkey)
    - [`GetFeatureVariable`](https://developers.kameleoon.com/feature-management-and-experimentation/web-sdks/go-sdk/#getfeaturevariable)
    - [`GetFeatureVariationVariables`](https://developers.kameleoon.com/feature-management-and-experimentation/web-sdks/go-sdk/#getfeaturevariationvariables)
* Changed `Data` types:
  * [`Browser`](https://developers.kameleoon.com/feature-management-and-experimentation/web-sdks/go-sdk/#browser):
    - Added constructor [`NewBrowser`](https://developers.kameleoon.com/feature-management-and-experimentation/web-sdks/go-sdk/#newbrowser)
    - Hid the `Type` and `Version` fields and replaced them with getter methods
  * [`Conversion`](https://developers.kameleoon.com/feature-management-and-experimentation/web-sdks/go-sdk/#conversion):
    - Added constructors [`NewConversion`](https://developers.kameleoon.com/feature-management-and-experimentation/web-sdks/go-sdk/#newconversion) and [`NewConversionWithRevenue`](https://developers.kameleoon.com/feature-management-and-experimentation/web-sdks/go-sdk/#newconversionwithrevenue)
    - Hid the `GoalId`, `Revenue`, and `Negative` fields and replaced them with getter methods
  * [`CustomData`](https://developers.kameleoon.com/feature-management-and-experimentation/web-sdks/go-sdk/#customdata):
    - Changed the data type of the `ID` field to `int`
    - Hid the `ID` field and replaced with a getter method
    - Renamed `GetValues` to `Values`
  * [`Device`](https://developers.kameleoon.com/feature-management-and-experimentation/web-sdks/go-sdk/#device):
    - Added constructor [`NewDevice`](https://developers.kameleoon.com/feature-management-and-experimentation/web-sdks/go-sdk/#newdevice)
    - Hid the `Type` field and replaced with a getter method
  * [`PageView`](https://developers.kameleoon.com/feature-management-and-experimentation/web-sdks/go-sdk/#pageview):
    - Added constructors [`NewPageView`](https://developers.kameleoon.com/feature-management-and-experimentation/web-sdks/go-sdk/#newpageview), [`NewPageViewWithTitle`](https://developers.kameleoon.com/feature-management-and-experimentation/web-sdks/go-sdk/#newpageviewwithtitle)
    - Hid the `URL`, `Title`, and `Referrers` fields and replaced with the getters
    - Changed `url` from optional to a required parameter
  * [`UserAgent`](https://developers.kameleoon.com/feature-management-and-experimentation/web-sdks/go-sdk/#useragent):
    - Added constructor [`NewUserAgent`](https://developers.kameleoon.com/feature-management-and-experimentation/web-sdks/go-sdk/#newuseragent)
    - Hid `Value` field and replaced with the getter
* Reworked cookies:
    - Removed `SetVisitorCode` method
    - Removed `ObtainVisitorCode` method
    - Removed parameter `topLevelDomain` from [`GetVisitorCode`](https://developers.kameleoon.com/feature-management-and-experimentation/web-sdks/go-sdk/#getvisitorcode). Instead, use the `top_level_domain` parameter in the [configuration](https://developers.kameleoon.com/feature-management-and-experimentation/web-sdks/go-sdk/#additional-configuration)
    - Made [`GetVisitorCode`](https://developers.kameleoon.com/feature-management-and-experimentation/web-sdks/go-sdk/#getvisitorcode) return `(string, error)` pair.
* Removed methods that were deprecated in 2.x versions:
    - `RetrieveDataFromRemoteSource`
* Removed visitor data max size:
    - Removed `visitor_data_max_size` configuration field
    - Removed `VisitorDataMaxSize` field from `Config`

### Features
* Added a [`SetLegalConsent`](https://developers.kameleoon.com/feature-management-and-experimentation/web-sdks/go-sdk/#setlegalconsent) method to determine the types of data Kameleoon includes in tracking requests. This helps you adhere to legal and regulatory requirements while responsibly managing visitor data. You can find more information in the [Consent management policy](https://help.kameleoon.com/consent-management-policy/).
* Implemented `KameleoonClientFactory`(https://developers.kameleoon.com/feature-management-and-experimentation/web-sdks/go-sdk/#kameleoonclientfactory) to manage `KameleoonClient` instances:
    - [`Create`](https://developers.kameleoon.com/feature-management-and-experimentation/web-sdks/go-sdk/create)
    - [`CreateFromFile`](https://developers.kameleoon.com/feature-management-and-experimentation/web-sdks/go-sdk/#createfromfile)
    - [`Forget`](https://developers.kameleoon.com/feature-management-and-experimentation/web-sdks/go-sdk/#forget)
* Added [`WaitInit`](https://developers.kameleoon.com/feature-management-and-experimentation/web-sdks/go-sdk/#waitinit) method to wait until the initialization process is completed
* Added new parameters for `KameleoonClientConfig`:
    - `SessionDuration` ([`session_duration`](https://developers.kameleoon.com/feature-management-and-experimentation/web-sdks/go-sdk#additional-configuration) configuration field accordingly)
    - `TopLevelDomain` field to `Config` ([`top_level_domain`](https://developers.kameleoon.com/feature-management-and-experimentation/web-sdks/go-sdk#additional-configuration) configuration field accordingly)

### Bug fixes
* Stability and performance improvements

## 2.3.1 - 2023-10-03
### Features
* Added support for older versions of the Go language. You can now use version 1.12 or later (previously, the minimum version was 1.16).

## 2.3.0 - 2023-09-12
### Features
* Added [`GetRemoteVisitorData`](https://developers.kameleoon.com/feature-management-and-experimentation/web-sdks/go-sdk/#GetRemoteVisitorData) method to fetch a visitor's remote data (with an optional capability to add the fetched data to the visitor).
### Bug fixes
* Stability and performance improvements

## 2.2.0 - 2023-08-20
* Stability and performance improvements

## 2.1.1 - 2023-06-28
* Added new conditions for targeting:
    - `Visitor Code`
    - `SDK Language`
    - [`Page Title & Page Url`](https://developers.kameleoon.com/feature-management-and-experimentation/web-sdks/go-sdk/#pageview)
    - [`Browser`](https://developers.kameleoon.com/feature-management-and-experimentation/web-sdks/go-sdk/#browser)
    - [`Device`](https://developers.kameleoon.com/feature-management-and-experimentation/web-sdks/go-sdk/#device)
    - [`Conversion`](https://developers.kameleoon.com/feature-management-and-experimentation/web-sdks/go-sdk/#trackconversion)

## 2.1.0 - 2023-04-05
* Added update campaigns and feature flag configurations instantaneously with Real-Time Streaming Architecture: [`Documentation`](https://developers.kameleoon.com/go-sdk.html#streaming) or [`Product Updates`](https://www.kameleoon.com/en/blog/real-time-streaming)
* Added a new methods:
    - [`OnUpdateConfiguration`](https://developers.kameleoon.com/go-sdk.html#onupdateconfiguration) to handle events when configuration data is updated in real time
    - [`GetEngineTrackingCode`](https://developers.kameleoon.com/go-sdk.html#GetEngineTrackingCode) which can be used to simplify utilization of hybrid mode
* Minor bug fixes

## 2.0.6 - 2023-03-24
* Minor bug fixing for `is among the values` operator for [`CustomData`](https://developers.kameleoon.com/go-sdk.html#customdata).
* Renaming of methods:
    - `RetrieveDataFromRemoteSource`-> [`GetRemoteData`](https://developers.kameleoon.com/go-sdk.html#GetRemoteData)

## 2.0.5 - 2023-03-13
* Added possibility for [`CustomData`](https://developers.kameleoon.com/go-sdk.html#customdata) to use variable argument list of values
* Fixed issue with [`TriggerExperiment`](https://developers.kameleoon.com/go-sdk.html#triggerexperiment) returning wrong error for client with no experiments
* Fixed issue with provided `Config` with not specified values; related to [initialization](https://developers.kameleoon.com/go-sdk.html#1-initialize-the-kameleoon-client)

## 2.0.4 - 2023-03-07
* [`GetVariationAssociatedData`](https://developers.kameleoon.com/go-sdk.html#getvariationassociateddata) Fixed. No need to unquote bytes upon obtaining anymore.

## 2.0.3 - 2023-02-27
* Minor bug fixing

## 2.0.2 - 2023-01-02
* Removed dependency on first version

## 2.0.1 - 2023-01-02
* Fixed issue with distribution of v2

## 2.0.0 - 2023-01-02
* Significantly improved configuration load time
* Added support for **Experiment** & **Exclusive Campaign** conditions. Related to [`TriggerExperiment`](https://developers.kameleoon.com/go-sdk.html#triggerexperiment)
* Renaming of methods:
    - `ActivateFeature`-> [`IsFeatureActive`](https://developers.kameleoon.com/go-sdk.html#IsFeatureActive)
    - `ErrNotActivated` -> `ErrNotAllocated`. Related to [`TriggerExperiment`](https://developers.kameleoon.com/go-sdk.html#triggerexperiment)
* Methods added for obtaining experiment and feature flag lists along with feature variables:
    - [`GetFeatureAllVariables`](https://developers.kameleoon.com/go-sdk.html#GetFeatureAllVariables)
    - [`GetFeatureList`](https://developers.kameleoon.com/go-sdk.html#GetFeatureList)
    - [`GetActiveFeatureListForVisitor`](https://developers.kameleoon.com/go-sdk.html#GetActiveFeatureListForVisitor)
    - [`GetExperimentList`](https://developers.kameleoon.com/go-sdk.html#GetExperimentList)
    - [`GetExperimentListForVisitor`](https://developers.kameleoon.com/go-sdk.html#GetExperimentListForVisitor)
* Added support of `is among the values` operator for Custom Data
* Added KameleoonData [`Device`](https://developers.kameleoon.com/go-sdk.html#device) data. Possible values are: **Phone**, **Tablet**, **Desktop**.
* Removed KameleoonData `Interest`

## 1.0.6 - 2022-04-12
* Added method for retrieving data from remote source: [`RetrieveDataFromRemoteSource`](https://developers.kameleoon.com/go-sdk.html#retrievedatafromremotesource)

## 1.0.5 - 2022-02-28
* Added support of multi-environment for feature flags, Related to [`ActivateFeature`](https://developers.kameleoon.com/go-sdk.html#activatefeature), [`GetFeatureVariable`](https://developers.kameleoon.com/go-sdk.html#getfeaturevariable)
* Added checking for status of site_code (Enable / Disable). Related to [`ActivateFeature`](https://developers.kameleoon.com/go-sdk.html#activatefeature), [`TriggerExperiment`](https://developers.kameleoon.com/go-sdk.html#triggerexperiment)

## 1.0.4 - 2021-12-10
* Added scheduling functionality for [`ActivateFeature`](https://developers.kameleoon.com/go-sdk.html#activatefeature)
* Added VisitorCodeNotValid exception when empty or exceeding the limit of 255 chars. Related to [`ActivateFeature`](https://developers.kameleoon.com/go-sdk.html#activatefeature), [`TriggerExperiment`](https://developers.kameleoon.com/go-sdk.html#triggerexperiment), [`AddData`](https://developers.kameleoon.com/go-sdk.html#adddata), [`TrackConversion`](https://developers.kameleoon.com/go-sdk.html#trackconversion), [`FlushVisitor`](https://developers.kameleoon.com/go-sdk.html#flush)

## 1.0.3 - 2021-12-06
* GraphQL API is using now instead of REST
* Improved SDK stability

## 1.0.2 - 2021-12-03
* Fixed issue with wrong bucketing. Related to [`TriggerExperiment`](https://developers.kameleoon.com/go-sdk.html#triggerexperiment)

## 1.0.1 - 2021-11-30
* Improved SDK stability

## 1.0.0 - 2021-06-24
* Added Fasthttp
