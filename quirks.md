PLOT TWIST
  service keys can be credhub references too!
  cloud controller fetches the values from credhub
  https://github.com/cloudfoundry/cloud_controller_ng/blob/65a75e6c97f49756df96e437e253f033415b2db1/app/collection_transformers/credhub_credential_populator.rb#L14
  https://github.com/cloudfoundry/cloud_controller_ng/blob/a732bdfe9c8013e69fd610f342b00c0ce384b1c8/spec/unit/controllers/services/service_keys_controller_spec.rb#L533
  https://github.com/cloudfoundry/cloud_controller_ng/blob/ea0bf3387a069395377a9df0ef04cd29a482eb9b/lib/credhub/client.rb#L10
  so there's another endpoint to sort out...

ANOTHER PLOT TWIST
  `cf curl /` returns the URL of credhub that's configured!
  it'll expose whatever we use internally, eek
  https://github.com/cloudfoundry/cloud_controller_ng/blob/99978117368f5b917b4caaf674ca9e647dc124e5/app/controllers/runtime/root_controller.rb#L50

SOLUTIONS?
* Add the `/api/v1/data?name=#{reference_name}&current=true` -> ['data'][0]['value'] endpoint to our CredHub-imitator, supporting the UAA client used by Cloud Controller [see https://github.com/cloudfoundry/cloud_controller_ng/blob/ea0bf3387a069395377a9df0ef04cd29a482eb9b/lib/credhub/client.rb#L10]
* Be doubly-sure the URL is internal and doesn't look like a public thing
* Actually have to configure a public credhub url so not a problem [see https://github.com/cloudfoundry/capi-release/blob/develop/jobs/cloud_controller_ng/spec#L1125-L1126]

THOUGHTS!
* Service keys can wait! It's only if someone tries to fetch a service key including `credhub-ref` that they'll even trigger a minor error. Right now they'd need to write a custom service broker to do that
* Knowing about this extra CredHub integration I feel a little less keen to support something other than CredHub. That fricking sucks because I don't want CredHub to have a future.
