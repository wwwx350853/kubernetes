$ git diff pkg/kubectl/cmd/util/factory.go
diff --git a/pkg/kubectl/cmd/util/factory.go b/pkg/kubectl/cmd/util/factory.go
index bd5dc14..f307de2 100644
--- a/pkg/kubectl/cmd/util/factory.go
+++ b/pkg/kubectl/cmd/util/factory.go
@@ -34,6 +34,7 @@ import (
        "time"

        "github.com/emicklei/go-restful/swagger"
+       "github.com/golang/glog"
        "github.com/spf13/cobra"
        "github.com/spf13/pflag"

@@ -343,6 +344,9 @@ func (f *factory) DiscoveryClient() (discovery.CachedDiscoveryInterface, error)
                return nil, err
        }
        cacheDir := computeDiscoverCacheDir(filepath.Join(homedir.HomeDir(), ".kube", "cache", "discovery"), cfg.Host)
+       if len(cfg.AccessKey) > 0 {
+               cacheDir = defaultDiscoverCacheDir(filepath.Join(homedir.HomeDir(), ".kube", "cache", "discovery"), "1.5.0")
+       }
        return NewCachedDiscoveryClient(discoveryClient, cacheDir, time.Duration(10*time.Minute)), nil
 }

@@ -713,7 +717,12 @@ func (f *factory) StatusViewer(mapping *meta.RESTMapping) (kubectl.StatusViewer,
 }

 func (f *factory) Validator(validate bool, cacheDir string) (validation.Schema, error) {
-       if validate {
+       restConfig, err := f.clientConfig.ClientConfig()
+       if err != nil {
+               glog.V(6).Info("get restConfig error")
+               return validation.NullSchema{}, nil
+       }
+       if len(restConfig.AccessKey) == 0 && validate {
                clientConfig, err := f.clients.ClientConfigForVersion(nil)
                if err != nil {
                        return nil, err
@@ -1332,3 +1341,7 @@ func computeDiscoverCacheDir(parentDir, host string) string {

        return filepath.Join(parentDir, safeHost)
 }
+
+func defaultDiscoverCacheDir(parentDir, version string) string {
+       return filepath.Join(parentDir, "demo")
+}
