package rwasm

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/suborbital/reactr/request"
	"github.com/suborbital/reactr/rt"
	"github.com/suborbital/vektor/vlog"
)

type testBody struct {
	Username string `json:"username"`
}

var sharedRT *rt.Reactr

func init() {
	// set a logger for rwasm to use
	UseLogger(vlog.Default(
		vlog.Level(vlog.LogLevelDebug),
	))

	// create a shared instance for some tests to use
	sharedRT = rt.New()

	if err := HandleBundleAtPath(sharedRT, "./testdata/runnables.wasm.zip"); err != nil {
		fmt.Println(errors.Wrap(err, "failed to AtHandleBundleAtPath"))
		return
	}
}

func TestWasmRunnerWithFetch(t *testing.T) {
	r := rt.New()

	// test a WASM module that is loaded directly instead of through the bundle
	doWasm := r.Handle("wasm", NewRunner("./testdata/fetch/fetch.wasm"))

	res, err := doWasm("https://1password.com").Then()
	if err != nil {
		t.Error(errors.Wrap(err, "failed to Then"))
		return
	}

	if len(res.([]byte)) < 100 {
		t.Errorf("expected 1password.com HTML, got %q", string(res.([]byte)))
	}

	if string(res.([]byte))[:100] != "<!doctype html><html lang=en data-language-url=/><head><meta charset=utf-8><meta name=viewport conte" {
		t.Errorf("expected 1password.com HTML, got %q", string(res.([]byte))[:100])
	}
}

func TestWasmRunnerWithFetchSwift(t *testing.T) {
	job := rt.NewJob("fetch-swift", "https://1password.com")

	res, err := sharedRT.Do(job).Then()
	if err != nil {
		t.Error(errors.Wrap(err, "failed to Then"))
		return
	}

	if string(res.([]byte))[:100] != "<!doctype html><html lang=en data-language-url=/><head><meta charset=utf-8><meta name=viewport conte" {
		t.Error(fmt.Errorf("expected 1password.com HTML, got %q", string(res.([]byte))[:100]))
	}
}

func TestWasmRunnerEchoSwift(t *testing.T) {
	job := rt.NewJob("hello-swift", "Connor")

	res, err := sharedRT.Do(job).Then()
	if err != nil {
		t.Error(errors.Wrap(err, "failed to Then"))
		return
	}

	if string(res.([]byte)) != "hello Connor" {
		t.Error(fmt.Errorf("hello Connor, got %s", string(res.([]byte))))
	}
}

func TestWasmRunnerWithRequest(t *testing.T) {
	r := rt.New()

	// using a Rust module
	doWasm := r.Handle("wasm", NewRunner("./testdata/log/log.wasm"))

	body := testBody{
		Username: "cohix",
	}

	bodyJSON, _ := json.Marshal(body)

	req := &request.CoordinatedRequest{
		Method: "GET",
		URL:    "/hello/world",
		ID:     uuid.New().String(),
		Body:   bodyJSON,
		State: map[string][]byte{
			"hello": []byte("what is up"),
		},
	}

	reqJSON, err := req.ToJSON()
	if err != nil {
		t.Error("failed to ToJSON", err)
	}

	res, err := doWasm(reqJSON).Then()
	if err != nil {
		t.Error(errors.Wrap(err, "failed to Then"))
		return
	}

	if string(res.([]byte)) != "hello what is up" {
		t.Error(fmt.Errorf("expected 'hello, what is up', got %s", string(res.([]byte))))
	}
}

func TestWasmRunnerSwift(t *testing.T) {
	body := testBody{
		Username: "cohix",
	}

	bodyJSON, _ := json.Marshal(body)

	req := &request.CoordinatedRequest{
		Method: "GET",
		URL:    "/hello/world",
		ID:     uuid.New().String(),
		Body:   bodyJSON,
		State: map[string][]byte{
			"hello": []byte("what is up"),
		},
	}

	reqJSON, err := req.ToJSON()
	if err != nil {
		t.Error("failed to ToJSON", err)
	}

	job := rt.NewJob("swift-log", reqJSON)

	res, err := sharedRT.Do(job).Then()
	if err != nil {
		t.Error(errors.Wrap(err, "failed to Then"))
		return
	}

	if string(res.([]byte)) != "hello what is up" {
		t.Error(fmt.Errorf("expected 'hello what is up', got %q", string(res.([]byte))))
	}
}

func TestWasmRunnerDataConversion(t *testing.T) {
	r := rt.New()

	doWasm := r.Handle("wasm", NewRunner("./testdata/hello-echo/hello-echo.wasm"))

	res, err := doWasm("my name is joe").Then()
	if err != nil {
		t.Error(errors.Wrap(err, "failed to Then"))
	}

	if string(res.([]byte)) != "hello my name is joe" {
		t.Error(fmt.Errorf("expected 'hello my name is joe', got %s", string(res.([]byte))))
	}
}

func TestWasmRunnerGroup(t *testing.T) {
	r := rt.New()

	doWasm := r.Handle("wasm", NewRunner("./testdata/hello-echo/hello-echo.wasm"))

	grp := rt.NewGroup()
	for i := 0; i < 50000; i++ {
		grp.Add(doWasm([]byte(fmt.Sprintf("world %d", i))))
	}

	if err := grp.Wait(); err != nil {
		t.Error(err)
	}
}

func TestWasmBundle(t *testing.T) {
	res, err := sharedRT.Do(rt.NewJob("hello-echo", []byte("wasmWorker!"))).Then()
	if err != nil {
		t.Error(errors.Wrap(err, "Then returned error"))
		return
	}

	if string(res.([]byte)) != "hello wasmWorker!" {
		t.Error(fmt.Errorf("expected 'hello wasmWorker!', got %s", string(res.([]byte))))
	}
}

const largeInput = `
Lorem ipsum dolor sit amet, consectetur adipiscing elit. Mauris vel ante a tellus consequat placerat at ut odio. Sed ut sagittis turpis. Sed ullamcorper in magna id posuere. Maecenas turpis nibh, vestibulum quis arcu vitae, commodo dignissim massa. Sed blandit, ligula ut euismod elementum, massa odio varius eros, vitae euismod eros ligula vel diam. Donec odio sapien, placerat vitae velit a, facilisis scelerisque libero. Suspendisse vel blandit ex, vitae cursus ex.

Curabitur sagittis elementum urna rhoncus tristique. Proin at dolor in arcu imperdiet mattis quis at leo. Fusce elementum ipsum diam, non molestie nisi elementum et. Nullam at nunc quis nibh tristique sollicitudin. Suspendisse ac enim felis. Integer ut ultricies nisl. Fusce nunc mi, finibus vitae lorem vitae, tincidunt auctor dolor. Sed viverra dolor eu nisl maximus, a finibus odio placerat. Orci varius natoque penatibus et magnis dis parturient montes, nascetur ridiculus mus. Morbi malesuada tortor nec diam tempor, euismod pharetra lorem imperdiet.

Proin vehicula vehicula molestie. Nullam feugiat lacus in consequat vehicula. Maecenas sed metus vitae risus sollicitudin scelerisque sed et enim. In sit amet turpis ante. Proin porta nibh nulla, non varius urna sodales vitae. Nullam turpis tellus, lacinia id elementum vitae, molestie id ipsum. Nullam tempor, elit at tempus vulputate, dolor magna pretium mi, vitae rhoncus odio enim a enim. In hac habitasse platea dictumst. Maecenas pretium, augue quis pulvinar gravida, tellus sem vehicula magna, in ornare dolor ante sed nibh.

Nunc eros mi, egestas at magna vitae, blandit tempus nisi. Integer viverra felis vel mauris molestie, et placerat est finibus. Quisque et nibh sit amet est sagittis luctus. Proin commodo risus at magna aliquet laoreet. Sed sollicitudin cursus massa bibendum vehicula. Fusce malesuada, felis eu facilisis faucibus, augue ante finibus lorem, nec mattis dui ex in urna. Sed vitae viverra metus. Quisque ut neque lacinia, malesuada justo nec, finibus justo. Pellentesque vestibulum lobortis augue, non tempor dolor. Vivamus congue eleifend condimentum. Proin tempor mollis mi ac ultrices. Vestibulum quis fermentum nibh, in posuere arcu.

Cras et sapien ex. In vel molestie orci, in suscipit nunc. Pellentesque habitant morbi tristique senectus et netus et malesuada fames ac turpis egestas. Proin ac nulla sit amet augue sagittis semper. Vivamus finibus massa nec efficitur posuere. Pellentesque auctor risus neque, porta venenatis nunc euismod vel. Donec sollicitudin mauris eget faucibus consectetur. Proin enim quam, gravida et consectetur vitae, iaculis a orci. Suspendisse placerat, massa at lobortis mollis, purus massa volutpat felis, non tristique quam orci sit amet ipsum. Pellentesque eget tincidunt magna, ut lobortis enim. Vestibulum varius mi imperdiet tortor venenatis maximus. Cras aliquam iaculis maximus. Proin tellus augue, feugiat a odio et, placerat lacinia lorem. Donec a metus est. Phasellus lobortis nibh ante. Integer eget libero blandit, vulputate purus sit amet, accumsan sem.

Praesent placerat nisi id dolor placerat ullamcorper. Fusce mattis nisi nec lectus vulputate, ut gravida lacus molestie. Morbi sollicitudin leo eros, non tempus ligula dictum a. Curabitur enim diam, tristique vel lacus eu, cursus hendrerit ipsum. Phasellus pharetra vel lectus at egestas. Donec suscipit risus sagittis rutrum condimentum. Maecenas eros libero, facilisis nec ligula ut, semper volutpat elit. Morbi in consectetur mi. Donec blandit accumsan luctus. Suspendisse potenti. Aliquam sit amet est tortor. Phasellus magna nibh, auctor at nunc sed, pharetra ultricies nisi. Nunc eleifend urna ac nibh dignissim, in molestie purus feugiat. Sed scelerisque luctus imperdiet.

Donec in libero maximus, porttitor ex vel, interdum eros. Duis ullamcorper odio ac dapibus fermentum. Proin dapibus non lorem sed dignissim. Aliquam et nibh et dolor rhoncus placerat in ac arcu. Aliquam id scelerisque nibh. Cras gravida et sem pretium interdum. Proin posuere nunc nulla, at vehicula massa porta quis. Morbi sollicitudin porttitor eros, vel cursus erat hendrerit vel. Quisque feugiat ipsum vel mauris tempus, non dictum enim mattis. Aliquam erat volutpat. Proin gravida, lacus sodales consectetur iaculis, justo lectus faucibus turpis, a elementum leo enim vitae leo. Praesent purus ex, egestas non orci ac, eleifend pellentesque dui. Donec fringilla congue est sit amet cursus. Curabitur viverra nec augue vitae tempus. Donec eu odio neque. Nunc facilisis sapien sit amet mollis pellentesque.

Morbi scelerisque faucibus ipsum, vitae pellentesque mi sollicitudin vel. Pellentesque tristique leo ac dolor pellentesque, sed cursus dolor facilisis. Integer imperdiet laoreet augue ut euismod. Donec sit amet nisl mauris. Nullam non ante vitae lectus pretium bibendum. Suspendisse rhoncus feugiat quam, non eleifend augue posuere sed. Etiam ut laoreet nisl, at consequat arcu. Donec leo risus, congue a accumsan non, pharetra id urna. Nam finibus erat vitae augue viverra sodales. Interdum et malesuada fames ac ante ipsum primis in faucibus.

Aenean vestibulum sodales odio eget dictum. Nunc non enim id leo tincidunt malesuada. Mauris viverra et quam at volutpat. Suspendisse dictum, magna eu volutpat pretium, sem ante ultricies felis, quis ultrices diam diam ut massa. Nullam feugiat molestie tincidunt. Donec porttitor vestibulum erat ac aliquam. Morbi consequat lectus nulla, ac feugiat nibh ultricies interdum. Mauris mi tellus, congue nec elit sit amet, gravida varius dui. Aenean euismod urna ut nibh egestas accumsan. Nunc ut mauris eu quam mattis tincidunt. In hac habitasse platea dictumst. Duis congue vitae nisl eu placerat. Suspendisse porta suscipit nisl, volutpat bibendum mauris porta eget. Pellentesque interdum tellus bibendum porttitor dignissim. Vivamus interdum mauris vel nibh eleifend auctor.

Nullam pulvinar sapien justo, eget consequat risus dictum quis. Vestibulum non blandit elit, et pretium purus. Nulla facilisi. Sed sit amet ipsum ac felis imperdiet luctus. Aliquam congue mauris sed purus molestie, eu tempus dolor pulvinar. Phasellus sit amet tincidunt nisl. Curabitur eget libero tincidunt ligula aliquam tincidunt non sed enim. Proin ac fermentum velit.

Fusce efficitur egestas purus quis rutrum. Donec hendrerit semper hendrerit. Suspendisse auctor id mauris vitae volutpat. Ut molestie cursus tempor. Proin nec auctor dui, eget fermentum dolor. Phasellus sit amet sem eget sem semper tincidunt a ac quam. Nullam at nulla placerat, vulputate lorem eu, ultricies mi. Praesent ac arcu mi. Integer hendrerit nulla vel posuere vulputate. Pellentesque bibendum ornare mi eget ornare. Nunc massa sapien, eleifend vitae venenatis a, pharetra eget ipsum. Suspendisse ullamcorper at risus at mattis. Aliquam pretium dolor in nulla semper, ut dictum sapien molestie. Morbi vulputate lorem lectus, eu suscipit urna porttitor a. Vestibulum id mi sed mauris congue eleifend.

Donec in magna eu dolor commodo accumsan. Quisque fermentum dolor ac metus bibendum placerat. Cras eu lacus lectus. Maecenas ac fermentum quam, ac convallis ipsum. Fusce faucibus vel ipsum vel commodo. Donec sit amet venenatis sapien. Etiam luctus augue non tellus pharetra auctor. Nulla nec nunc non velit tincidunt ultrices. Duis consequat nec justo ut mattis. Vivamus egestas euismod quam vitae dictum. Donec quis lacus libero. Nunc quis porta mi. Duis rutrum, elit nec rhoncus fermentum, dui nulla ultricies sem, eu tincidunt purus est eu erat.

Donec at tellus vestibulum, auctor nisl nec, accumsan mi. Nam quis faucibus nisl. Proin accumsan mollis neque, et eleifend eros cursus eget. Nullam efficitur at diam id rutrum. Proin et nisi at erat venenatis pulvinar. Mauris eget lectus fermentum, consectetur sapien a, feugiat leo. Duis quis metus quis sem feugiat tincidunt in pharetra ex. Donec mattis risus facilisis risus vehicula, a efficitur lorem tristique. Cras a dignissim purus, vitae maximus risus. Nullam eu lectus at eros vehicula pharetra eu sed metus. Pellentesque quis diam est. Donec fermentum sapien tempor congue vulputate. Suspendisse orci nisl, mollis non vulputate vitae, maximus vel ex. Pellentesque aliquet bibendum dui in interdum. Duis venenatis purus a arcu cursus scelerisque. Etiam porttitor molestie quam, ac porta nisi dapibus suscipit.

Ut imperdiet lectus condimentum turpis vestibulum tempor. Suspendisse sit amet ligula elit. Phasellus nec nunc consequat, consectetur turpis nec, semper arcu. Donec quis ex ut ex cursus dictum ac nec enim. Aenean commodo ipsum id est posuere, et tincidunt metus lacinia. Orci varius natoque penatibus et magnis dis parturient montes, nascetur ridiculus mus. Nulla in aliquet mauris. Pellentesque ultricies nisl in ligula gravida suscipit. Donec dui mi, viverra in convallis commodo, ornare eget tortor. Sed cursus tellus id felis faucibus venenatis. Integer et orci eu urna congue viverra ac at nulla.

Fusce eros arcu, maximus eu tempus vitae, ultricies non lectus. Proin hendrerit ultricies augue, aliquam ultrices dui facilisis a. Aenean pulvinar pulvinar sapien quis porttitor. Curabitur elit augue, iaculis quis ex vitae, sollicitudin vulputate sem. Etiam vehicula odio vitae ante pharetra, id pulvinar lectus fringilla. Suspendisse at justo at arcu egestas tincidunt eu at est. Pellentesque congue maximus enim, sit amet ornare dolor.

Ut at ultrices lorem. In convallis rutrum tortor, nec scelerisque magna gravida non. Aenean massa lorem, maximus a sodales sed, elementum nec mauris. Quisque posuere euismod scelerisque. Maecenas venenatis risus ante, nec molestie ex eleifend quis. Donec sed leo quis odio faucibus finibus sed nec odio. Curabitur semper orci enim, ut volutpat eros suscipit et. Proin consectetur dui nisl. Etiam id risus pretium nulla scelerisque sollicitudin.

Cras pulvinar nisl eu odio congue finibus. Interdum et malesuada fames ac ante ipsum primis in faucibus. Sed vitae bibendum mauris. Ut nulla ligula, pharetra ac sagittis efficitur, tincidunt maximus ligula. Aliquam sit amet mauris in felis elementum condimentum at molestie lacus. Proin leo nunc, convallis quis sapien nec, dignissim hendrerit purus. Interdum et malesuada fames ac ante ipsum primis in faucibus. Suspendisse fringilla tellus nec porttitor ullamcorper. Aenean ac eros est. Proin suscipit orci nec enim tempus finibus. Sed vehicula egestas tellus quis feugiat. Nullam blandit dolor vitae facilisis auctor. Sed et massa vel nisi pellentesque interdum ac nec libero. Morbi a justo sed massa aliquam pharetra. Nunc efficitur lectus nisi, viverra cursus lacus eleifend a.

Fusce arcu elit, accumsan a dolor id, rhoncus auctor odio. Aenean tempor mi dui, eu pretium ipsum euismod ut. Curabitur porta rutrum ligula, sed ultrices tellus. In ut pulvinar justo. In hac habitasse platea dictumst. Donec accumsan nisl sed dolor tincidunt feugiat. Nam mi nisl, interdum a magna non, auctor posuere metus. Lorem ipsum dolor sit amet, consectetur adipiscing elit. Etiam tincidunt sollicitudin turpis ac tristique. Donec venenatis interdum nibh, lobortis malesuada quam euismod at. Integer elementum ligula quam, vitae cursus nunc lobortis sed. Vivamus ultricies urna turpis, vitae aliquam urna consectetur in. In varius mattis arcu, quis euismod dui placerat sed. Fusce sed mollis mi, vitae imperdiet purus. Interdum et malesuada fames ac ante ipsum primis in faucibus. Aliquam ullamcorper metus a est bibendum, a commodo odio dictum.

Maecenas imperdiet ornare placerat. Pellentesque sed nibh pellentesque, consectetur urna ac, varius ante. Aenean ac viverra tortor, vel mattis neque. Pellentesque molestie est eget sagittis feugiat. Lorem ipsum dolor sit amet, consectetur adipiscing elit. Maecenas euismod urna sit amet porttitor congue. Phasellus at ultricies justo. Vestibulum consequat purus et pretium sagittis. Nulla facilisi. Mauris consectetur ante dolor. Integer vel semper massa, sit amet tempor ligula. Fusce imperdiet orci quis magna faucibus blandit. Morbi laoreet mi diam, ac facilisis ante condimentum non. Proin nec justo nec neque mollis dapibus id a mi. Aliquam diam odio, posuere sit amet ipsum ut, finibus maximus arcu.

Fusce elit orci, congue nec enim quis, consequat porta quam. Etiam lacinia elit eros, vel pretium mi commodo sed. Aenean rhoncus porttitor tortor, convallis interdum nulla venenatis et. Fusce augue lorem, accumsan eget blandit vitae, facilisis eu sapien. Phasellus accumsan imperdiet mattis. Sed pharetra sem ut auctor gravida. Quisque placerat, dolor nec feugiat varius, tortor justo aliquam dolor, non pulvinar turpis neque ut eros. Integer dictum varius purus, ac vulputate elit facilisis in. In aliquam laoreet turpis vel mollis. Nullam fermentum neque laoreet ipsum luctus porttitor. Suspendisse posuere dictum nisl.

Vestibulum ullamcorper posuere euismod. Nunc vulputate justo et augue tincidunt facilisis. Maecenas at tincidunt turpis. Praesent ultricies varius lacus, ac fermentum mi sagittis sed. Maecenas a volutpat massa, sit amet blandit velit. Sed condimentum semper tristique. Aliquam dictum nulla mi, et bibendum libero euismod at. Vestibulum purus sem, fermentum in sagittis sit amet, semper sit amet est. Phasellus justo lacus, mattis ut scelerisque a, gravida eu nibh. Aenean vulputate turpis sed enim condimentum convallis. Quisque congue velit id leo euismod, in facilisis mauris rhoncus.

Morbi ut dui rhoncus enim faucibus rutrum. Donec consequat suscipit lorem, eget convallis est bibendum sit amet. Vestibulum mi felis, accumsan vel varius nec, accumsan at lorem. Praesent malesuada tincidunt faucibus. Mauris consectetur erat vel purus venenatis rutrum sed id dui. Curabitur vehicula placerat odio at rutrum. Phasellus volutpat tempus dolor, vel molestie dui ultricies eu. Pellentesque consectetur arcu vel mollis lacinia. Nulla accumsan, urna ac tempor tempor, mauris leo accumsan lectus, in aliquam orci nisl sit amet leo. In porta sollicitudin ultricies. Sed bibendum ultricies turpis, et laoreet leo fringilla convallis. Morbi commodo luctus nulla, vel efficitur nunc. Etiam tincidunt lorem velit, id mollis nibh fringilla in. Nunc vestibulum eu orci ac dignissim. Nullam pellentesque metus enim, in viverra mauris eleifend et. Interdum et malesuada fames ac ante ipsum primis in faucibus.

Nunc orci sem, dictum eget neque quis, viverra molestie nibh. Sed mattis aliquet mi nec cursus. Curabitur tempor sapien ac nisl blandit, sit amet finibus mi porttitor. Vestibulum nec lobortis velit. Suspendisse in sem felis. Proin accumsan nisi metus, et faucibus lacus ultricies vitae. Pellentesque blandit risus ac diam vehicula, id lobortis mauris dictum. Donec rutrum facilisis risus ac vestibulum. Vestibulum volutpat elit vitae quam pretium imperdiet. Vivamus porta mi luctus varius finibus. Aenean ut consectetur sapien, quis cursus sapien. Proin eu metus iaculis, pellentesque quam a, sodales justo. Donec nunc ante, bibendum ac urna in, iaculis mollis quam. Ut euismod pellentesque pretium. Donec non commodo neque. Proin pulvinar risus et enim tempor, ut scelerisque nunc suscipit.

Quisque vestibulum ex velit, vel malesuada felis porta vitae. Pellentesque habitant morbi tristique senectus et netus et malesuada fames ac turpis egestas. Phasellus congue aliquam commodo. Duis vel velit semper, vestibulum dui in, tempus arcu. Vestibulum varius vel lectus dictum dictum. Ut non mattis quam, ac sollicitudin leo. Ut efficitur enim eu mollis tincidunt. Nulla facilisi. Pellentesque vehicula interdum lectus, a cursus eros laoreet quis. Duis lobortis tortor sit amet ultrices gravida. Fusce pellentesque sem in quam dignissim varius. Quisque pulvinar massa massa, consequat blandit elit auctor eu. Phasellus eu massa lorem.

Etiam maximus dolor condimentum vehicula cursus. Praesent risus dolor, ullamcorper et volutpat a, eleifend nec justo. Sed consequat posuere odio id blandit. Curabitur vel egestas quam, in tincidunt tellus. Donec fringilla elit vitae leo vehicula porttitor. Pellentesque vel justo nec nisl congue pulvinar. Suspendisse blandit justo dapibus mauris efficitur vehicula. Donec eu fermentum lectus. Aliquam in consequat ipsum, imperdiet ullamcorper risus. Cras nisi nisl, pretium eget turpis quis, tempor egestas nisl. Aenean in sapien id nisl facilisis faucibus. Aenean ut nulla et nisi vehicula ultricies non ut nibh. Phasellus pharetra orci nec varius dictum.

Cras malesuada elit in mauris rutrum, eu aliquet mi varius. Aenean rhoncus, mi vitae malesuada accumsan, ipsum libero egestas risus, eget bibendum turpis ex vitae nisi. Maecenas augue risus, condimentum at quam at, dignissim tincidunt mauris. Vestibulum nec quam accumsan, efficitur sem eu, dignissim est. Nam volutpat, dolor molestie egestas congue, ligula elit facilisis ante, sit amet maximus lorem metus a elit. Aliquam vel dapibus velit, vitae pulvinar massa. Suspendisse sit amet ligula aliquet, tempus lorem nec, blandit lectus. Suspendisse vitae leo et velit cursus auctor ut iaculis massa. Suspendisse potenti. Aenean vitae vehicula lacus. Nunc augue leo, interdum faucibus mi quis, auctor euismod quam. Sed eleifend diam arcu, eu rutrum ligula laoreet in. Etiam non ipsum arcu. Fusce dignissim neque a suscipit aliquet.

Integer sit amet leo ipsum. Sed consectetur facilisis purus id porta. Mauris efficitur lacus nunc, sit amet interdum dui viverra et. In massa ipsum, semper eu neque a, facilisis vulputate purus. Phasellus commodo neque diam, a dignissim nibh vehicula at. Vivamus tincidunt vestibulum sapien, pretium gravida odio interdum a. Duis dapibus tempor lacus, in varius nibh pulvinar vel. Morbi vestibulum tempor tortor, non imperdiet ante molestie in. Phasellus sit amet sapien nec tortor semper consectetur id sit amet tortor. Sed mi mauris, euismod id fermentum ut, rutrum in enim. Integer vitae sem nisl. Donec rhoncus lacus non mauris lobortis, quis rhoncus nulla dictum. Vivamus velit metus, iaculis sed pretium egestas, consectetur at nisi. Quisque vitae volutpat ligula, a molestie nisi. Etiam sagittis maximus ligula in interdum. Maecenas facilisis ante non lectus sagittis, eget tempor nisi lobortis.

Nam et efficitur nibh. Suspendisse at tortor dictum, dignissim elit ut, imperdiet purus. Sed in elit ex. Mauris vehicula sodales arcu sit amet tempor. Curabitur vel ante neque. Nullam maximus commodo velit, eu blandit nisl cursus sit amet. Nullam sed ex eu purus volutpat gravida et quis enim.

Donec et tristique metus. In euismod lectus malesuada auctor rhoncus. Pellentesque accumsan risus et accumsan posuere. Sed non turpis rhoncus magna gravida maximus. Curabitur id sem malesuada, dignissim nisi vitae, congue felis. Aenean ipsum velit, lobortis mattis sollicitudin sit amet, euismod ac purus. Nam ornare tincidunt diam at cursus. Donec ac risus sed ante rutrum hendrerit sed eget massa. Nulla ut erat magna. Quisque et nisl et diam ullamcorper vulputate. Aenean porttitor enim massa, in tempus mi dictum vel. Cras efficitur at ipsum a volutpat. Integer lacinia vulputate ornare. Sed fermentum et sem quis elementum.

Curabitur egestas, nibh quis condimentum faucibus, neque nunc convallis nunc, sed condimentum lorem nisl ac est. Aliquam luctus odio diam, molestie convallis sapien mollis et. Etiam sollicitudin, dolor id varius laoreet, nisl sem condimentum dui, nec facilisis lacus est et urna. Nullam placerat elit vel ligula eleifend ornare. Maecenas odio erat, fringilla nec luctus quis, elementum nec felis. Cras faucibus at leo et aliquam. Vivamus finibus sapien non iaculis sodales. Etiam quis velit nunc. Proin euismod molestie molestie. Nulla ut ipsum quis eros egestas luctus.

Nam fermentum, quam ut molestie pulvinar, lorem neque gravida orci, in luctus nibh felis aliquam est. Fusce bibendum odio nec lobortis porttitor. Nulla tempor nisl ac massa sagittis, sit amet finibus elit eleifend. Maecenas condimentum orci non nisl rutrum lobortis. Integer facilisis, massa vitae ullamcorper iaculis, tellus tortor maximus nunc, imperdiet porttitor arcu lorem at orci. Integer in condimentum ante. Phasellus aliquam justo sed bibendum posuere. In mattis efficitur felis et accumsan. Praesent tincidunt nisi vel metus efficitur blandit. Morbi facilisis, mi id aliquam bibendum, velit metus venenatis augue, quis venenatis elit tellus sit amet est. Sed vel tempus elit. Mauris sit amet massa finibus, facilisis diam vel, bibendum sapien. Integer massa turpis, laoreet non enim vel, pellentesque ullamcorper felis. Donec accumsan elit non purus ullamcorper euismod. Nulla viverra nibh libero, eget commodo elit interdum vel.

In consectetur orci libero. Praesent consequat rhoncus fermentum. Nam ut elit lacus. Nulla facilisi. Pellentesque sollicitudin est ut facilisis commodo. Nunc tincidunt quam id lectus accumsan efficitur. Cras sed hendrerit magna. In at felis elit. Vestibulum lacinia vel ipsum vitae pulvinar. Aliquam arcu ipsum, condimentum vitae luctus sed, condimentum in magna.

Ut id tincidunt metus, ut aliquet eros. Donec iaculis sollicitudin odio, semper aliquam sem fermentum a. Mauris venenatis consectetur sem, vel placerat erat faucibus vel. Phasellus eleifend justo semper iaculis sodales. Orci varius natoque penatibus et magnis dis parturient montes, nascetur ridiculus mus. Mauris auctor sodales massa. Etiam sit amet consectetur sapien. Aenean id orci pharetra, rutrum orci non, luctus lacus. Vivamus lacinia arcu nec lectus vulputate vestibulum. Nullam nibh neque, pellentesque at purus eget, laoreet bibendum velit. Maecenas finibus facilisis eros, eu tristique lectus accumsan et. Donec et lectus non dui dignissim pretium. Etiam eget tincidunt massa. Donec orci diam, sagittis eget porta id, rhoncus a ex.

Morbi ultrices mi ornare justo facilisis fermentum. Pellentesque auctor pellentesque est id tristique. Nullam efficitur, lectus et scelerisque consectetur, ligula leo viverra orci, vitae efficitur nisi nisl ut metus. Mauris maximus imperdiet nisi. Vestibulum volutpat mollis magna, id rutrum massa imperdiet id. Quisque vitae sem aliquet, efficitur elit non, eleifend odio. Nulla facilisi. Phasellus dapibus nisl eget purus eleifend, at malesuada dui aliquam. In rutrum ante id urna sagittis, ac tempor lacus blandit. Sed sed turpis dignissim, rutrum lectus ac, dictum nisi. Nam molestie consectetur felis hendrerit congue. Nulla facilisi. Donec in dui vitae tellus pharetra ornare ut ut augue. Morbi augue libero, feugiat non finibus rutrum, euismod maximus odio. Nunc placerat pretium pulvinar. Quisque ut nibh condimentum, ornare ex eu, varius nunc.

Phasellus nec ex libero. Integer sed felis urna. Cras tempus nisi tellus, sed volutpat libero aliquet et. Fusce sit amet aliquam sapien, id tincidunt lacus. Mauris convallis vulputate eros, at molestie nisl sollicitudin in. Duis tincidunt vestibulum molestie. Morbi sem justo, viverra in venenatis eget, varius non est. Vivamus blandit viverra nisi non finibus. Integer iaculis luctus euismod. Vivamus eu gravida massa, eu auctor ipsum. Nunc venenatis ut elit ut laoreet. Pellentesque cursus at lorem eget posuere. Integer finibus, elit pretium euismod luctus, ipsum lorem congue leo, eu pellentesque nisi nulla et sapien. Curabitur iaculis, tortor nec dignissim vehicula, augue nulla feugiat felis, quis tempus massa nulla sit amet velit. Sed laoreet erat ultricies, volutpat leo id, sollicitudin sapien.

Fusce in nibh libero. Donec dapibus enim ac mollis auctor. Etiam et ex quis libero auctor hendrerit. Fusce dui dui, ornare eu aliquam eget, varius vel urna. Aliquam fermentum posuere ultricies. Phasellus laoreet dui quam, in rutrum mauris eleifend non. Nulla in eros elit. Ut suscipit a mauris quis vestibulum. Mauris vehicula nisi felis, a viverra leo rutrum non. Donec porttitor mi vitae eros molestie, nec ullamcorper libero scelerisque. Cras et facilisis tellus, quis gravida orci. Mauris eleifend lorem sit amet dignissim ornare.

Proin malesuada viverra metus, non bibendum mi sollicitudin ac. Suspendisse vel metus iaculis, tristique mauris in, placerat elit. Duis urna enim, venenatis id magna ut, bibendum ultrices est. Cras non elit non ex auctor mollis. Donec vel commodo ex. Donec dapibus sit amet nunc id faucibus. Nunc imperdiet pulvinar mi, at commodo mauris maximus et. Pellentesque auctor, mi ac euismod tincidunt, mauris augue molestie magna, varius ullamcorper enim sem et massa. Vivamus sit amet quam elit. Aenean nec nisl accumsan, convallis tellus eget, hendrerit elit.

Cras eu odio ut eros tristique hendrerit id at orci. Vivamus suscipit viverra tempor. Donec sit amet purus venenatis, iaculis tellus non, euismod justo. Integer massa dui, sollicitudin quis vulputate eu, luctus sit amet mi. Cras auctor interdum feugiat. Vivamus scelerisque euismod semper. Mauris et condimentum sapien. Vestibulum commodo sapien ac magna accumsan, eget rutrum orci commodo. Ut ornare tortor sit amet nunc commodo, id ullamcorper sapien lacinia. Nullam eget tellus quam. Duis lobortis a nunc et dapibus. Phasellus in auctor quam, non pretium nisi. Pellentesque luctus molestie eros.

Pellentesque habitant morbi tristique senectus et netus et malesuada fames ac turpis egestas. Morbi fermentum libero eget lorem aliquam cursus. Curabitur varius at libero ut semper. Phasellus ut felis enim. Donec at augue quis tortor sollicitudin vehicula id eget mi. Praesent eu ligula dictum, auctor nunc sit amet, consequat ante. Nulla posuere lectus eu sollicitudin porta. Nullam tempus augue eget justo interdum, non cursus libero porttitor. Sed convallis suscipit dui sed aliquet. Maecenas placerat blandit orci, eu facilisis massa dignissim vitae. Proin in rutrum urna, pulvinar ultrices dui. Nulla posuere viverra lorem, at elementum augue consectetur et. Quisque nisi erat, viverra quis placerat non, sodales quis ante.

Morbi et urna sed justo malesuada tempor vitae et odio. Duis ut sem hendrerit, sagittis sapien vel, porttitor risus. Aenean semper mauris et orci ullamcorper, a fringilla erat maximus. Vivamus tincidunt urna et nulla rutrum mattis. Aliquam porta non quam in viverra. Suspendisse risus ante, suscipit et lorem vel, rutrum laoreet est. Nullam at nisi et ante hendrerit varius id non sapien. In ut venenatis odio. Nam id volutpat odio. Sed non tellus et ipsum laoreet tempor nec et urna. Donec posuere tincidunt aliquam. Vestibulum vehicula tincidunt nibh, vel commodo nisi rutrum sit amet. Nulla facilisi.

Cras accumsan pellentesque nisl quis aliquet. Fusce fringilla semper diam sed condimentum. Nam sed vulputate nunc, at fermentum tortor. Donec non ornare nisi. Sed nec ligula in est sagittis fermentum. Cras congue magna lacinia purus dignissim, sed placerat augue luctus. Nunc et dui et tortor imperdiet aliquam sed et odio. Maecenas egestas ultrices pellentesque.

Duis vulputate nulla at tempor mollis. Quisque at porta velit. Curabitur vel diam nec felis ultricies consectetur ut sed enim. Mauris ut neque in quam ullamcorper convallis. Proin erat ipsum, eleifend quis mi in, ultrices imperdiet ipsum. Nunc efficitur ornare dui, ac elementum diam tincidunt in. Curabitur imperdiet tellus in purus tristique, quis fermentum justo convallis. Aliquam id convallis erat. Mauris congue faucibus est, sit amet lacinia purus tristique vitae. Vivamus sit amet elit quis tortor laoreet eleifend. Nullam pulvinar blandit aliquam. Morbi nec nisl tincidunt, pulvinar dui a, euismod sem. Curabitur eget tortor quis purus pulvinar vehicula et eget elit. Suspendisse aliquet, erat ac volutpat suscipit, purus nisi ornare urna, eu vulputate arcu turpis eget velit. Donec condimentum tortor ac leo porttitor, vitae ultricies ligula pretium.

Pellentesque dignissim, elit eu rutrum blandit, erat nunc commodo ligula, non blandit lacus tellus at tellus. Etiam luctus, tellus et congue convallis, lectus mi pretium elit, ac luctus dolor est vitae justo. Maecenas quis purus commodo, fringilla ante id, ullamcorper erat. Vestibulum rutrum augue eu turpis dictum accumsan. Praesent pharetra, nisi at scelerisque pellentesque, arcu justo rutrum nunc, vel commodo dolor sapien id tellus. Morbi tortor metus, eleifend et elementum vitae, pharetra in nisi. Aenean in erat id enim pulvinar semper non ut urna. Praesent scelerisque pharetra eros, vitae sagittis massa sollicitudin vitae. Aliquam erat volutpat. Proin tempor turpis eros, non pulvinar dolor mollis vitae. Integer sapien justo, facilisis quis justo eu, feugiat blandit massa. Vivamus urna massa, consectetur vitae arcu id, pretium pretium metus. Suspendisse viverra bibendum lacus, eu finibus magna finibus in. Mauris quis odio eu tortor mattis efficitur. Proin neque mauris, pellentesque eget velit et, cursus mattis metus. Mauris et risus in metus pretium dapibus.

Sed at fringilla ante. Donec scelerisque volutpat elementum. Aliquam at laoreet nisl. Nunc porta urna enim, et lacinia lacus suscipit ullamcorper. Praesent in mi sollicitudin, vestibulum sapien quis, malesuada lorem. Sed eget tellus volutpat, facilisis magna ac, elementum arcu. Aenean eget diam viverra, vestibulum dolor at, congue urna. Pellentesque viverra justo sed ante pellentesque, eu rutrum massa eleifend. Sed viverra orci at mi gravida, dapibus vehicula diam cursus. Phasellus a mattis ex, fringilla dapibus nisi.

Lorem ipsum dolor sit amet, consectetur adipiscing elit. Vivamus id ex suscipit, commodo metus non, elementum nunc. Nullam molestie purus quis commodo euismod. Cras lobortis velit non eleifend viverra. Nam bibendum vel ligula et dapibus. Duis id odio eu enim tempus rhoncus. Donec vitae purus in velit vulputate imperdiet. Mauris et lacus et arcu hendrerit feugiat eu eget purus. Praesent lobortis vulputate massa, vel viverra ipsum porttitor eu. Suspendisse euismod nisi at vestibulum dictum. Etiam justo purus, pharetra ut elit vel, condimentum sodales libero. Mauris eget faucibus diam. Proin nec odio et urna venenatis venenatis. Donec quam nulla, pulvinar bibendum lectus a, eleifend imperdiet lectus. Suspendisse potenti.

Fusce quis libero vel leo congue scelerisque vestibulum non dolor. Suspendisse convallis lacus euismod sem elementum condimentum. Vivamus ultrices enim ut arcu volutpat sodales. Sed volutpat sapien massa, lacinia vestibulum magna tristique nec. Proin eget justo id magna aliquam molestie quis vel dolor. Suspendisse consectetur ligula ut tellus posuere ultrices. Morbi in velit imperdiet justo consequat consequat at id massa. Donec sagittis sit amet libero sit amet dignissim. Praesent scelerisque mauris quis sem ultricies molestie. Interdum et malesuada fames ac ante ipsum primis in faucibus. Vestibulum vitae sem tempus nisl malesuada rhoncus.

Proin placerat dolor non sodales pellentesque. Pellentesque tellus magna, eleifend eget molestie et, maximus eget lacus. Nulla facilisi. Maecenas velit lectus, blandit eu mi vel, ornare lobortis nunc. Cras maximus vehicula felis. Praesent fringilla elit at accumsan condimentum. Phasellus euismod leo ac nulla consequat venenatis. Nullam id tellus bibendum, consequat tortor ac, aliquam lorem. Morbi consequat erat purus, non fringilla nibh pretium non.

Aliquam erat volutpat. Aenean vulputate orci leo, ac laoreet erat elementum id. Praesent maximus ex in dui auctor, vel ultricies est elementum. Phasellus quis arcu nisl. Nam lobortis tincidunt mauris id suscipit. Mauris urna orci, tincidunt quis ante at, lobortis iaculis dui. Nulla ultrices augue eget urna sodales tempor. Sed elementum, nibh ac fringilla condimentum, eros purus gravida nunc, et luctus ligula libero et sem.

Integer nunc orci, dignissim sed sapien sit amet, auctor congue arcu. In ac dolor sit amet diam scelerisque auctor. Duis vehicula, orci et mattis mattis, est turpis porta quam, eget lobortis enim eros nec felis. Donec vulputate dui sit amet ligula porttitor consequat. Ut suscipit magna sed urna condimentum tempus. Etiam quis risus rutrum, gravida quam eget, aliquet erat. Aliquam tempor ut urna nec luctus. Vivamus non nisi ornare, sodales lectus id, commodo dolor. Mauris neque ex, commodo in neque in, sagittis molestie purus.

Donec accumsan luctus nibh vel fermentum. Nunc consequat turpis eget elementum porta. Ut blandit ligula non aliquet congue. Donec pellentesque consectetur magna, a viverra turpis tempor accumsan. Phasellus semper arcu non arcu malesuada, in rutrum lacus facilisis. Morbi tempor tristique faucibus. Nunc gravida porttitor sodales. Sed viverra malesuada ex. Nam id sollicitudin sem. Integer commodo non urna et suscipit. Sed eu luctus velit. Vestibulum condimentum orci sem, non tristique nunc interdum nec. Sed lobortis massa tellus, a lobortis urna volutpat eget. Quisque a ligula eget erat convallis porttitor non non dui. Sed iaculis tincidunt lacus id scelerisque.

Ut id ipsum orci. Vestibulum finibus orci urna, sed bibendum quam convallis nec. Praesent dapibus tempor nulla ut maximus. Nam tortor felis, commodo ut dapibus non, pretium at lorem. Praesent pretium a augue vel porta. Fusce eu semper nisl, quis laoreet urna. Donec viverra sit amet arcu a vulputate. Nunc volutpat est in ipsum faucibus, eu eleifend enim tincidunt. In ullamcorper congue velit, gravida dictum sapien volutpat a. Praesent ullamcorper libero eu placerat cursus. Vivamus laoreet metus non pellentesque euismod. Nullam accumsan, risus a auctor congue, nisi mi viverra risus, non rutrum metus velit ac metus. Vestibulum ante ipsum primis in faucibus orci luctus et ultrices posuere cubilia Curae; Suspendisse potenti.

Sed leo tellus, sagittis in augue sit amet, mattis sagittis ante. Sed lorem ante, sodales non elementum et, auctor at turpis. Nulla ac venenatis diam. In id euismod mi. Integer quis convallis libero, id dignissim magna. Aenean mollis sodales mi, at tincidunt felis lacinia nec. Donec ut turpis eu orci fermentum ultrices. Maecenas volutpat gravida lectus at ultrices. Mauris eu leo sit amet metus commodo aliquet. Integer semper justo in est suscipit, sed dictum metus sollicitudin.

In cursus, erat nec blandit ultricies, nunc leo posuere tellus, ut sodales lorem odio ornare lorem. Phasellus suscipit nibh sed augue consequat lacinia. Pellentesque in consequat dui. Nunc dignissim, mi ut viverra accumsan, libero nisl aliquet ex, et rhoncus dui tortor et leo. Praesent condimentum vehicula tempor. Suspendisse sit amet turpis aliquam dui mattis semper. Maecenas congue ipsum quis turpis ultricies, ac convallis erat imperdiet. Etiam a hendrerit lorem. Vestibulum non urna in nibh sodales ornare.

Class aptent taciti sociosqu ad litora torquent per conubia nostra, per inceptos himenaeos. Morbi id interdum orci, et tempus neque. Mauris volutpat, libero vel gravida auctor, nibh arcu finibus felis, vel fermentum mauris ante eget arcu. Etiam consequat magna tellus, ut vulputate magna porttitor vitae. Fusce tristique volutpat suscipit. Donec lacus nisl, mollis ut euismod vehicula, molestie eget erat. Proin eget magna posuere, pretium erat eget, venenatis odio. Vivamus suscipit erat bibendum, suscipit ante sit amet, interdum mauris.

Duis gravida, magna sit amet porttitor iaculis, erat elit dignissim nulla, vel tincidunt libero libero sed risus. Suspendisse varius tristique nisl vitae pulvinar. Vestibulum tincidunt tincidunt dictum. Aliquam erat volutpat. Duis suscipit dolor quam, vel elementum augue congue ut. Donec diam tortor, suscipit semper quam ut, tincidunt hendrerit nisi. Donec ac sapien id tortor pharetra tempus sit amet ut nulla. Maecenas odio felis, semper at dictum sed, rhoncus id orci. Nunc eu risus quis justo accumsan feugiat ut ut velit. Vestibulum eget convallis mauris, in pulvinar erat. Etiam fermentum tortor convallis aliquam convallis. Nam ut arcu ac est sagittis laoreet. Nulla non auctor tellus. Praesent aliquam dapibus euismod. Cras feugiat lorem eu ex viverra auctor. Ut dignissim luctus euismod.

Vivamus lobortis rhoncus pellentesque. Integer varius mauris sit amet purus scelerisque, non scelerisque risus porttitor. Etiam vulputate, sem et tempor luctus, nunc sem lacinia purus, ultricies vulputate justo orci in lorem. Maecenas luctus ac diam nec imperdiet. Pellentesque arcu tortor, mollis ac est ac, posuere gravida lacus. Suspendisse non diam at dui gravida molestie. Cras blandit semper odio nec congue. Aliquam imperdiet pellentesque blandit. Aliquam ligula tortor, rhoncus eu ligula vitae, consectetur feugiat enim. Praesent in facilisis est.

Curabitur ac ipsum justo. Nunc sit amet velit ut arcu semper facilisis sed eu ex. Fusce eget placerat est. Nulla consequat tellus neque, nec hendrerit ligula pulvinar sit amet. Nunc dictum tempor cursus. Sed pellentesque dui metus, et placerat mauris dignissim vel. Mauris sit amet libero a eros volutpat accumsan vel vitae ipsum. Quisque sed dapibus magna. Nunc commodo elit sit amet tristique aliquet.

Vestibulum molestie vitae ex nec tincidunt. Aliquam a malesuada est, sagittis feugiat orci. Maecenas blandit tincidunt dictum. Etiam ac consequat arcu. Sed molestie enim orci. Duis porttitor tincidunt massa, nec iaculis quam semper gravida. Suspendisse porttitor risus erat, at iaculis ex placerat nec. Nullam nec augue vel est tristique vestibulum. Donec scelerisque velit turpis, elementum consequat mauris commodo in. Pellentesque habitant morbi tristique senectus et netus et malesuada fames ac turpis egestas.

Cras ac porta arcu. Nam lectus ipsum, rhoncus ac nulla quis, lacinia cursus ex. Fusce sed leo sit amet velit rutrum porttitor at eu neque. Pellentesque habitant morbi tristique senectus et netus et malesuada fames ac turpis egestas. Curabitur vel ante risus. Proin convallis urna augue. Etiam dignissim lorem eget vehicula volutpat. Nam egestas, nibh sit amet sagittis facilisis, urna tellus pellentesque neque, et pellentesque risus justo ac dui. Suspendisse scelerisque porta volutpat. Duis laoreet quis eros vitae auctor. Nullam eu orci est. Etiam vel sodales velit. Nullam non velit metus. Nullam posuere laoreet lectus et hendrerit. Quisque accumsan nibh lacus, luctus lobortis massa condimentum non.

Curabitur neque dolor, elementum vel quam vel, lobortis efficitur magna. Mauris et elit pulvinar, luctus dolor in, tincidunt augue. Morbi sed tellus eget nibh ullamcorper aliquam eget sed nisi. Pellentesque urna ante, consequat et mattis sit amet, tincidunt sed mi. In laoreet porttitor eros, vitae venenatis urna euismod eu. Sed porttitor nunc vitae commodo tristique. Quisque sed nisl tempus, consequat orci in, feugiat magna. Sed pretium elit vitae dui auctor, vitae tincidunt mi tristique. Nunc consectetur ante a metus ornare, eget venenatis ante volutpat. Mauris at viverra urna, a blandit justo. Nam euismod sapien velit, porttitor cursus nulla ultricies eu. Aenean tristique aliquet sem, quis aliquet sapien fringilla quis. Proin varius ex quis purus dapibus tempor. Vestibulum ipsum justo, semper vitae condimentum sit amet, porta a libero. Duis facilisis sapien odio, sed elementum massa commodo non.

Pellentesque in augue at justo eleifend mattis. Nullam id rhoncus arcu. In ultrices ornare massa, at convallis elit. Duis sit amet mi tempus, dignissim ante lacinia, efficitur sapien. Donec vulputate dui id est tristique porta. Sed et felis non lacus posuere ornare. Sed eget diam et velit varius tincidunt vel ut urna. Curabitur tempus tempus magna in tincidunt.

Nunc ac ornare urna. Praesent a imperdiet urna, at consectetur mi. Sed consequat dui a fermentum dictum. Praesent ultricies, erat in ullamcorper interdum, risus ligula viverra libero, at consectetur lectus quam sit amet ex. Fusce non sollicitudin lacus. Integer tempus pulvinar purus nec viverra. In dictum velit velit, at ultricies lectus placerat ut. Praesent aliquet, metus eu ultricies consequat, massa mauris hendrerit augue, sollicitudin viverra odio nibh consequat elit. Morbi at orci ut lectus tempus elementum. Duis interdum egestas nulla non ultrices. Praesent laoreet congue nulla ac porta.

Etiam lacinia diam nibh, a venenatis ipsum bibendum sit amet. Maecenas vitae iaculis libero, id faucibus ipsum. Sed placerat justo a ipsum fringilla, et tincidunt metus accumsan. Sed venenatis blandit mi. Cras non velit nec est pharetra hendrerit at a ligula. Nunc tempor eu velit in dignissim. Fusce et molestie magna. Nulla vel turpis egestas, egestas sapien quis, molestie odio. Ut velit sem, aliquet in neque ac, semper sollicitudin lectus.

Nulla auctor, velit ut condimentum sagittis, lacus ipsum semper velit, non tempus ligula ante faucibus justo. Vivamus id dignissim massa. Morbi vitae nunc non justo vehicula consectetur. Quisque ut magna efficitur, hendrerit augue ac, malesuada est. Duis sit amet semper ex, nec cursus odio. Duis commodo erat tortor, ut dapibus ante tempus et. Vestibulum suscipit, ipsum ut luctus rhoncus, est nibh tempor ante, quis tincidunt quam ligula sed nulla. Proin gravida nulla semper consectetur volutpat. Aenean eleifend porttitor feugiat.

Duis a ipsum nec enim sodales congue. Nullam id metus eu sapien venenatis efficitur id id tellus. Nullam id bibendum dolor. Morbi sagittis commodo odio scelerisque rhoncus. Maecenas ultrices, sem ac placerat pulvinar, massa quam dignissim felis, eu congue ligula felis in neque. Sed sagittis dapibus semper. Fusce quis sem leo. Etiam odio justo, tincidunt sed elementum maximus, ornare ac libero. Maecenas ultricies purus est, feugiat volutpat tortor faucibus quis. Interdum et malesuada fames ac ante ipsum primis in faucibus. Sed ac dapibus purus. Maecenas non dui id lacus dapibus condimentum ac vel orci. Morbi ipsum magna, ornare vitae vulputate ut, pretium eget mi. Praesent placerat tortor vitae urna tristique mattis. Aliquam velit enim, interdum eu arcu vitae, aliquam sollicitudin mauris.

Nunc laoreet, nisl at sodales aliquet, nunc metus fringilla risus, ut finibus augue neque in ante. Integer rutrum neque lacinia lectus dictum fringilla. Donec commodo massa libero, vitae consectetur orci congue in. Nulla sed pharetra tortor, ac pellentesque dolor. Suspendisse finibus, mi eget malesuada tincidunt, tellus justo gravida lectus, id mattis nisi ante sed diam. Donec pretium ornare porta. Morbi aliquam id neque vel vehicula. Pellentesque habitant morbi tristique senectus et netus et malesuada fames ac turpis egestas. Vivamus quis nisi sapien. Lorem ipsum dolor sit amet, consectetur adipiscing elit. Suspendisse a maximus mauris. Quisque quis orci cursus, hendrerit libero euismod, accumsan neque. Vestibulum laoreet velit id nulla aliquet iaculis.

Curabitur odio purus, molestie in ante ac, pharetra finibus dui. Nullam nec lacus rhoncus, venenatis mi vitae, interdum mauris. Praesent tincidunt mauris quis convallis elementum. Phasellus euismod ac nisl nec interdum. Nam faucibus, est vitae varius pulvinar, justo erat tempor neque, quis sollicitudin magna eros eu tortor. Nunc consectetur, lorem nec tincidunt maximus, magna urna auctor augue, euismod gravida leo metus sit amet erat. Etiam vitae nibh porta, lacinia lacus vel, vehicula leo. Sed bibendum odio auctor, scelerisque massa a, accumsan eros. Integer eu varius ex. Aenean non dapibus sem. Nulla blandit interdum lacus, nec efficitur nibh vestibulum sollicitudin. Proin in elit purus. Mauris eu laoreet lectus. Aenean in diam mi. Nulla sed rutrum urna, vitae ultricies nisi. Proin pretium lobortis venenatis.

Integer et nulla nec ante sodales consectetur a eu augue. Cras ut justo sed felis convallis volutpat eu ac lectus. Fusce eleifend, leo eu congue blandit, arcu nulla rhoncus tellus, non lacinia dolor nulla eget sem. In sollicitudin maximus libero sit amet sagittis. Sed efficitur sem sit amet leo aliquam ullamcorper. Curabitur suscipit nisi sed nisi volutpat, consequat bibendum lorem rutrum. Aenean a nisl vel mauris aliquam feugiat non nec ante. Nullam eget est dui. Pellentesque sollicitudin massa felis, nec vulputate lectus aliquam et. In eget enim vitae nulla iaculis faucibus ac nec nisl. Nulla tristique metus eget efficitur vulputate. Nunc sed diam pharetra, ultrices risus non, imperdiet metus.

Morbi iaculis magna non velit varius, sed ultrices massa pulvinar. Duis erat libero, pharetra vitae faucibus eu, pulvinar ut purus. Integer ut porta leo, a egestas tellus. Aenean tempus, mi et facilisis ultrices, ligula sapien placerat risus, et tempus nisi orci ac lorem. Donec non iaculis ante, id tincidunt mauris. Class aptent taciti sociosqu ad litora torquent per conubia nostra, per inceptos himenaeos. Aliquam auctor tellus augue, id tempus eros rhoncus vel. Aenean nec dapibus sem, quis fermentum dui. Sed tortor enim, faucibus eu odio id, bibendum pretium urna. Nullam accumsan, urna ac laoreet dictum, ex libero luctus neque, sit amet venenatis quam justo sit amet sapien.

Curabitur sit amet risus est. Praesent a massa iaculis, elementum enim et, rhoncus elit. Morbi accumsan commodo condimentum. Donec pellentesque dignissim augue, id lobortis nisi molestie in. Nulla commodo viverra arcu ut vestibulum. Vestibulum condimentum orci arcu, at sollicitudin odio tempor sit amet. Suspendisse ac massa sagittis, scelerisque tortor vel, feugiat mauris. Vestibulum et neque imperdiet nulla hendrerit euismod non non sapien.

Morbi sed sagittis libero. Nam eget rhoncus sapien, eu porttitor quam. Vivamus suscipit augue ut justo commodo consectetur. Cras orci lorem, commodo vel leo ut, malesuada finibus enim. Cras in facilisis lectus. Nullam at dignissim dolor. Cras eget varius odio. Nullam varius, lectus sodales mollis lacinia, turpis diam congue dui, sed sagittis nisi magna vulputate magna. Suspendisse quis arcu mi.

Nullam luctus eros ex, efficitur feugiat diam tristique sit amet. Sed in blandit tellus. Aliquam erat volutpat. Duis vel egestas erat. Duis vel orci condimentum, mollis mauris ut, fringilla eros. Nam et consequat odio. Ut porttitor, eros id tempor fermentum, nunc libero pretium ex, eget elementum est risus vitae est. Sed egestas, sem ac maximus lacinia, nulla velit rhoncus neque, non dignissim tortor tortor a felis. Nulla in metus mi. Curabitur tincidunt, tellus eget commodo placerat, orci lectus pretium purus, id mollis ligula nibh viverra neque. Suspendisse potenti. Nulla facilisi. Cras pharetra facilisis erat consequat consectetur. Donec nec nisl rhoncus, vulputate nunc quis, gravida enim. Donec luctus ut nisl sit amet facilisis. Cras vitae augue mauris.

Maecenas ac lorem nec tortor mattis convallis eget ut dolor. Pellentesque sed nunc molestie, maximus nisl sit amet, laoreet mi. Sed mollis mauris non finibus pulvinar. Praesent tincidunt, diam vitae faucibus luctus, arcu magna rhoncus sem, non placerat ante enim id quam. Nam ultrices scelerisque nunc, sit amet aliquam nisl pretium quis. Integer convallis dolor vitae risus scelerisque pellentesque. Nullam tincidunt augue et ipsum bibendum, vel vehicula dolor tristique. Maecenas pulvinar lacus sed dui vestibulum, ac feugiat felis facilisis. Mauris nec dolor sed dolor elementum aliquam et at massa. Donec porta neque nec euismod consequat. Lorem ipsum dolor sit amet, consectetur adipiscing elit. Praesent iaculis sodales neque eu hendrerit. Morbi quis dolor justo. Suspendisse porttitor, purus sed convallis pulvinar, nisl nisi mollis nisl, tincidunt lobortis dui massa ut turpis.

Donec convallis turpis et ultrices vestibulum. Nam scelerisque mauris lacus, eget pellentesque diam fermentum non. In fermentum eros est, non faucibus risus feugiat et. Etiam vitae felis eleifend, dignissim ligula vitae, accumsan ante. Phasellus mi arcu, dapibus vel porttitor vehicula, luctus non neque. Nunc aliquam lectus et hendrerit dictum. Interdum et malesuada fames ac ante ipsum primis in faucibus. Duis urna augue, euismod eget ipsum id, tincidunt mollis magna. Curabitur blandit, arcu id lacinia porttitor, enim libero tempor mauris, finibus tincidunt quam ex nec dolor. Ut lacinia porttitor nibh et sodales. Proin vel molestie justo. Integer mollis nibh sed libero aliquam faucibus. Integer faucibus lectus eu nunc lacinia, in pretium purus luctus. Pellentesque quis semper risus. Donec vitae velit velit. Vivamus venenatis, ligula nec vestibulum vehicula, tellus felis lacinia diam, at iaculis velit massa rhoncus erat.

Aliquam semper sed massa ut ornare. Fusce laoreet condimentum tortor, a iaculis tortor rhoncus mattis. Nulla sagittis tristique dui quis suscipit. Nulla eget nunc risus. Integer semper consequat placerat. Duis molestie nulla in leo tincidunt venenatis. Sed nec aliquet neque. Vivamus venenatis tellus eget eros hendrerit consequat. Aenean dapibus sem et pulvinar porta. Suspendisse sed suscipit quam. Mauris eget neque eget turpis fermentum accumsan vel in lectus. Duis vitae ipsum sapien.

Nam tincidunt mi quam, in ullamcorper nulla vehicula nec. Suspendisse luctus consequat lorem. Fusce commodo libero nulla, nec tempus risus mollis vitae. Phasellus eu aliquam elit. Sed in rhoncus magna. Cras at eros volutpat, faucibus purus finibus, sodales est. Nunc nisi nisi, ullamcorper eu venenatis et, consectetur vitae nibh. Donec elementum vitae ipsum eget vestibulum. Sed nec commodo quam, sit amet pharetra ante. Aliquam eget nisl in risus pretium lacinia. Suspendisse ac felis id leo vestibulum semper vitae vitae sem. Ut a placerat magna. Maecenas ac aliquam augue, nec placerat sapien. Pellentesque accumsan nulla sit amet elit porttitor hendrerit. Integer facilisis ligula et mauris dictum, ut sollicitudin lorem congue. Cras mattis laoreet quam eu placerat.

Nam quis faucibus augue, id dignissim sapien. Mauris vitae consequat ex. Etiam eu commodo elit. Ut commodo eros quis urna posuere maximus. Suspendisse potenti. In sit amet tortor sed nulla aliquam ultrices vitae vitae risus. Integer in tincidunt erat, id hendrerit neque. Praesent tempor mauris sed purus accumsan vestibulum. Nulla convallis lobortis mi, id viverra tellus. Sed ut ante ac sapien consectetur vulputate eget eget nisi. Nulla ac tincidunt mauris, at condimentum ante.

Integer iaculis ornare sodales. Mauris tempus nunc eget velit blandit, non facilisis lectus pellentesque. Cras sem turpis, hendrerit a semper condimentum, varius vitae nulla. Nullam maximus dapibus fringilla. Mauris vitae odio sit amet mi ultricies eleifend sit amet eu tortor. Phasellus faucibus massa vel libero luctus maximus. Curabitur non convallis nisl. Nullam nec mauris at ex suscipit vestibulum et a nibh. Nullam bibendum elit egestas ligula faucibus, eget interdum risus rhoncus. Mauris enim tortor, tincidunt quis laoreet ac, sagittis at mi. Nunc pellentesque tincidunt purus, aliquet semper purus rutrum quis. Suspendisse fringilla feugiat ipsum euismod volutpat. Nullam nisl enim, rutrum nec diam vel, viverra congue sem. Phasellus tincidunt, mi a auctor faucibus, lectus ex consectetur purus, id elementum turpis sapien nec mi. Etiam vitae diam et purus sollicitudin egestas. Nunc sollicitudin quam ultricies, faucibus libero ac, pharetra ex.

Nam ultricies, lacus at sagittis ornare, orci urna elementum massa, ut porta nunc lectus id massa. Class aptent taciti sociosqu ad litora torquent per conubia nostra, per inceptos himenaeos. Duis placerat ligula sit amet turpis luctus, et consequat odio condimentum. Donec auctor metus et tortor posuere convallis. Vivamus eget ultrices ligula. Etiam suscipit ornare purus, eget mollis ex vestibulum nec. Etiam feugiat fermentum tempus. Aenean blandit nibh sit amet purus pretium lobortis. Aenean imperdiet imperdiet dolor, at lobortis ligula finibus non. Nunc diam libero, laoreet at enim ut, semper sodales mauris. Donec mi justo, scelerisque sed cursus eget, blandit vitae lacus. Duis interdum pellentesque enim vulputate gravida. In imperdiet, velit ut volutpat euismod, quam arcu vestibulum lacus, at dapibus ante mauris id ante. Proin condimentum nisl ut diam egestas egestas a ut neque. Pellentesque rhoncus blandit convallis. Quisque magna nisi, auctor gravida congue ac, ornare vitae quam.

Duis gravida finibus lorem vitae elementum. Vestibulum velit tortor, venenatis at orci vitae, aliquam maximus nisl. Cras dictum dictum tempus. Proin dapibus neque ac velit fringilla dictum. Fusce fermentum semper ipsum, non pretium sapien varius vitae. Pellentesque hendrerit eu enim sed venenatis. Sed eu augue at mauris volutpat pharetra sed vitae diam. Integer nec leo in nulla mattis efficitur vel at nibh. Praesent auctor mauris id dapibus malesuada.

Donec ultrices lacus id pellentesque ultricies. Fusce viverra mi tincidunt, vulputate lacus eu, porta felis. Sed vehicula vitae quam in aliquam. Mauris a ligula in turpis tempus consequat. Ut ornare vulputate libero, in tempus nisi imperdiet sit amet. Nulla semper, ipsum id ullamcorper interdum, felis mi vulputate nibh, rhoncus hendrerit augue augue a magna. Etiam lorem velit, posuere quis viverra ac, facilisis non augue. Aliquam eros urna, vulputate hendrerit nisl nec, posuere ullamcorper sem. Vestibulum ante ipsum primis in faucibus orci luctus et ultrices posuere cubilia Curae; Fusce molestie auctor nisl, et viverra erat maximus quis. Donec faucibus eget ligula sit amet bibendum. Fusce semper ligula ac sem rhoncus porta. Donec mauris orci, pretium sed convallis at, auctor sed arcu. Duis vitae justo nisl. Donec a consequat neque.

Suspendisse libero leo, varius ut pellentesque vestibulum, consequat ac risus. Etiam mauris mi, condimentum ut sem malesuada, elementum sagittis ipsum. Pellentesque a rutrum ex, id fermentum turpis. Vestibulum non venenatis mi. Nulla tellus lorem, dictum convallis ultrices nec, consequat eu lorem. Maecenas scelerisque, tellus quis dictum vehicula, nibh lacus porta felis, fringilla dictum orci mauris a orci. Fusce vel condimentum mi. Nulla pulvinar bibendum bibendum. Etiam ultrices, neque sit amet laoreet pellentesque, mi orci tristique ex, ac scelerisque risus enim quis tellus. Nullam fringilla ante sed urna convallis ornare. Suspendisse velit sem, egestas vel arcu quis, viverra ultrices enim. Nulla a felis scelerisque, gravida ligula eget, lobortis sapien. Curabitur magna mi, venenatis eget convallis id, molestie ut augue. Nunc interdum urna sed ante vestibulum facilisis.

Mauris dictum risus erat, et egestas eros sagittis vel. Aenean at dapibus felis. Phasellus lacinia in ex non semper. Morbi bibendum ante vitae est elementum, sit amet pharetra nulla congue. Suspendisse quis quam tristique, porttitor lectus ac, vulputate velit. Vivamus in neque volutpat, suscipit libero in, tincidunt nisi. Lorem ipsum dolor sit amet, consectetur adipiscing elit. Aenean sit amet augue viverra, vulputate diam quis, rutrum mauris. Pellentesque habitant morbi tristique senectus et netus et malesuada fames ac turpis egestas.

Praesent in dui nunc. Etiam congue, erat ut pulvinar pellentesque, nunc eros volutpat sapien, vel volutpat odio eros id ante. Donec rhoncus orci turpis, id suscipit tellus pulvinar luctus. Phasellus euismod eros ut felis interdum commodo. Fusce volutpat odio nisi, eget aliquet metus pretium vel. Nunc tempor rhoncus lacus et ultrices. Vivamus rutrum ante leo, scelerisque cursus sapien placerat et. Donec suscipit feugiat risus ac fermentum. Integer at tempor felis. Mauris lobortis risus vitae augue accumsan dapibus. Fusce nec ex nec sapien consectetur euismod. Nullam et sem non libero aliquet iaculis. Maecenas facilisis sapien ut est rhoncus, sodales porta sapien mattis.

Sed nec ullamcorper odio, eget tempus lectus. Pellentesque egestas massa et auctor luctus. Nam lacinia ultrices mi id congue. Pellentesque a posuere nisi, in porttitor ipsum. Nulla egestas porta auctor. Duis posuere purus vitae urna scelerisque, vitae tempus ligula luctus. Vestibulum dignissim lorem est, ut elementum nisi auctor ac. Vestibulum vitae feugiat justo.

In tempus tincidunt ullamcorper. Etiam nibh urna, maximus in enim quis, tristique efficitur mi. Ut eu augue eget eros porttitor eleifend et ac enim. Donec et risus tellus. Vestibulum ante ipsum primis in faucibus orci luctus et ultrices posuere cubilia Curae; Nullam ut metus nunc. Donec nec pharetra sapien. Suspendisse potenti. Etiam ex ex, bibendum rhoncus rutrum eu, suscipit auctor diam. Sed mauris ligula, faucibus et tincidunt et, ullamcorper at orci. Nullam vel accumsan sapien. Nam lobortis, augue in cursus ornare, nunc est sollicitudin velit, sed iaculis erat eros at diam. Aenean iaculis orci sed diam rhoncus bibendum.

Aenean at sem eget elit varius volutpat quis nec turpis. Quisque mollis velit sit amet velit aliquet, ut sollicitudin ligula maximus. Fusce sodales erat id libero luctus tincidunt. Aenean euismod ornare turpis, id dapibus velit imperdiet in. Phasellus turpis elit, ornare eget tellus sit amet, posuere consectetur dolor. Donec ac diam condimentum, posuere leo eu, cursus nisi. Curabitur aliquam ipsum mi, ut laoreet augue vehicula quis. Cras vel dui ornare, finibus tellus in, pretium nibh. Praesent lacus neque, aliquam commodo placerat varius, tincidunt vitae mi. Fusce non dolor ut orci laoreet elementum vel vitae eros. Etiam eget ex nec odio porta sodales et eu ante. Aliquam imperdiet libero quis augue tincidunt, vel hendrerit metus suscipit. Phasellus est enim, aliquet eget tristique sit amet, facilisis sit amet augue. Praesent hendrerit lacinia urna sed pellentesque.

Sed venenatis egestas nunc in scelerisque. Etiam sed libero erat. Praesent laoreet rutrum nulla, id lobortis nibh porta vitae. Integer sapien nisl, placerat vitae feugiat in, luctus non neque. Donec felis metus, elementum eget elementum non, aliquam et leo. Vivamus felis libero, consectetur et mollis dictum, dignissim a elit. Quisque ante tellus, efficitur sit amet maximus at, dictum at dolor. Cras vehicula hendrerit sapien, eu volutpat enim finibus eu. Cras venenatis leo quis massa commodo, sit amet mollis metus porta. Suspendisse elementum in eros vitae sodales. Ut semper semper tortor, nec fermentum turpis laoreet vitae.

Praesent sit amet lacus a urna condimentum tincidunt. In pulvinar lobortis iaculis. Cras non odio eu libero accumsan sodales. Ut elementum ornare dui, vel ultrices nulla faucibus nec. Fusce dignissim tortor iaculis vehicula congue. Nulla sapien magna, dapibus sed scelerisque dictum, ullamcorper eu eros. Suspendisse lobortis dolor arcu.

Duis aliquam interdum dui eu bibendum. Suspendisse sit amet odio nisl. In id dui metus. Praesent a aliquet lacus. Interdum et malesuada fames ac ante ipsum primis in faucibus. Fusce lacinia nisl est, eu dictum lectus tempor sit amet. Cras faucibus lacinia lacus, nec vestibulum urna rutrum nec. Nam pulvinar, leo vitae ullamcorper eleifend, metus velit porttitor ipsum, ac iaculis sem lectus nec massa. Donec vel tellus et elit viverra euismod eget nec lacus. Etiam ligula orci, sagittis a turpis nec, finibus lacinia dolor. Sed in dignissim lacus, eget sagittis enim. Mauris lorem massa, feugiat quis lectus sed, posuere luctus ante. Nulla malesuada porta nisl. Duis vitae scelerisque felis. Donec sed leo et risus pretium pulvinar sed id purus.

Nam euismod, tellus a ultrices gravida, libero magna consectetur quam, ut mattis urna quam vitae metus. Fusce eleifend varius tristique. Morbi pharetra nibh urna, in condimentum velit ultrices ut. Duis ac rhoncus tortor, et ultricies ex. Sed congue, elit id facilisis lacinia, tortor erat commodo nisi, non malesuada dui nunc a erat. Donec sagittis dolor id vulputate tincidunt. Mauris vehicula lacus est, eget placerat tortor semper vitae. Mauris pellentesque rhoncus erat nec ornare.

Nunc auctor lectus id nunc egestas dapibus. Curabitur porta diam quis risus lacinia placerat. Aliquam erat volutpat. Nulla facilisi. Cras erat nibh, molestie nec lacus quis, dictum cursus tellus. Lorem ipsum dolor sit amet, consectetur adipiscing elit. Maecenas ultrices, erat facilisis maximus finibus, libero dui venenatis dolor, eget dictum nulla metus eget nisl. Nam quam ante, venenatis ac laoreet quis, varius nec nulla. Etiam convallis sollicitudin velit ac commodo. Aliquam auctor, leo ac efficitur iaculis, ante metus iaculis dolor, sit amet fringilla sapien risus in nisi. Pellentesque posuere volutpat convallis. Sed ullamcorper velit dui, sit amet sollicitudin magna pharetra vitae. Morbi congue enim sit amet velit tempor finibus. Nunc ultrices lacinia justo, eu feugiat tortor interdum et.

Ut rutrum commodo consectetur. Suspendisse ut vestibulum felis, vel porttitor nibh. Praesent vel orci posuere, ullamcorper arcu eget, congue nibh. Nam egestas volutpat egestas. Nullam mattis massa orci, et lacinia mauris lobortis aliquam. Integer viverra ipsum vel dignissim vehicula. Vestibulum id turpis ligula. Suspendisse justo est, dignissim id blandit dapibus, laoreet quis lorem. Praesent pharetra mi vel luctus accumsan. In congue volutpat lorem non dignissim.

Proin at gravida diam. Vestibulum vehicula convallis magna, quis varius justo finibus vitae. Etiam et luctus sapien. Proin dictum lacus ex, sed pulvinar leo ultrices sed. Sed sit amet leo sapien. Ut felis leo, scelerisque quis arcu id, consequat aliquet ligula. Cras est diam, convallis et pharetra et, iaculis consequat mi. Suspendisse metus felis, euismod ut scelerisque eu, rutrum eget sapien. Sed lacinia egestas odio, ac consectetur tellus ullamcorper vel. Orci varius natoque penatibus et magnis dis parturient montes, nascetur ridiculus mus. Aenean auctor dolor in interdum pulvinar. Vestibulum ante ipsum primis in faucibus orci luctus et ultrices posuere cubilia Curae; Nam ut nibh ac metus finibus pretium. Quisque lacinia arcu ac purus gravida laoreet. Praesent auctor magna sit amet turpis varius, eu blandit nisl interdum.

Suspendisse nec finibus augue, et finibus ligula. Phasellus ut orci eros. Nulla congue sed justo a porta. Curabitur est diam, scelerisque vel laoreet non, malesuada at erat. Aliquam eu viverra libero. Pellentesque habitant morbi tristique senectus et netus et malesuada fames ac turpis egestas. Suspendisse fringilla vel ex non ultrices. Nunc interdum tincidunt scelerisque. Donec augue mauris, ultricies eu consequat eu, placerat quis sem. Integer placerat viverra ex laoreet pharetra. Etiam non facilisis odio. Donec porta viverra mi, sed fermentum odio fermentum ut.

Vestibulum mattis ante et est commodo, vitae condimentum sapien viverra. Suspendisse eget velit in tellus varius accumsan nec non purus. Duis semper rutrum arcu non vehicula. Vivamus congue sem non ultricies rhoncus. Integer ultricies ipsum sapien, vitae porta enim volutpat nec. Nulla id mauris neque. Orci varius natoque penatibus et magnis dis parturient montes, nascetur ridiculus mus. Vestibulum quis arcu pellentesque, pretium lorem non, semper ligula. Morbi eu tincidunt lorem. Duis placerat pharetra hendrerit. Nulla ut viverra orci, vel auctor diam. Aliquam ultricies eros sed odio eleifend, ac mollis libero tempus. Quisque venenatis arcu sit amet tortor ultricies, vitae bibendum mauris sagittis. Phasellus vel dolor ut libero placerat condimentum. Suspendisse odio augue, sodales ut odio sed, ultrices faucibus neque.

Integer feugiat elit a diam rhoncus iaculis. Nam dolor purus, tristique et lectus vitae, pharetra sodales justo. Donec facilisis mauris quis lorem venenatis blandit. Proin nec lectus nec tortor consequat varius lobortis id est. Ut tortor risus, mattis sit amet massa ut, aliquam rutrum elit. Duis sodales elementum dui eu dapibus. Nam nec tempus justo. Nulla feugiat iaculis odio sed tempus. Vivamus a interdum elit, id malesuada velit. Ut rhoncus risus lorem, a ultrices lorem consectetur at. In et sapien at urna ullamcorper dapibus a non enim. Integer ac sapien ex.

Praesent sit amet nisi ligula. Nulla vitae quam eu nisl fermentum pulvinar vitae non metus. Etiam ultrices quam ex, eget tempor ante mattis et. Interdum et malesuada fames ac ante ipsum primis in faucibus. Vivamus vitae mauris quis justo convallis tincidunt eget id justo. Sed vulputate commodo diam ac suscipit. Vestibulum ante ipsum primis in faucibus orci luctus et ultrices posuere cubilia Curae; Integer eu turpis vitae nibh vestibulum porttitor ut at orci.

Vivamus vehicula leo mauris, at venenatis augue consequat in. Curabitur id dui nec eros rutrum laoreet. Donec dictum laoreet sapien. Donec rutrum dignissim luctus. Curabitur ante sem, sagittis a leo sed, placerat efficitur nisi. Duis a ligula sit amet arcu tincidunt sagittis. Praesent purus nisl, viverra sit amet rutrum et, consectetur commodo enim. Vivamus lacinia ex in dui fringilla, eget lacinia lacus ornare. Donec nec dui non magna tristique dignissim volutpat.
Generated 99 paragraphs, 9511 words, 64000 bytes of Lorem Ipsum
help@lipsum.com
`

func TestWasmLargeData(t *testing.T) {
	r := rt.New()

	doWasm := r.Handle("wasm", NewRunner("./testdata/hello-echo/hello-echo.wasm"))

	res := doWasm([]byte(largeInput))

	result, err := res.Then()
	if err != nil {
		t.Error(errors.Wrap(err, "failed to Then for large input"))
	}

	if len(string(result.([]byte))) < 64000 {
		t.Error(fmt.Errorf("large input job . too small, got %d", len(string(result.([]byte)))))
	}

	if string(result.([]byte)) != fmt.Sprintf("hello %s", largeInput) {
		t.Error(fmt.Errorf("large input result did not match"))
	}
}

func TestWasmLargeDataGroup(t *testing.T) {
	r := rt.New()

	doWasm := r.Handle("wasm", NewRunner("./testdata/hello-echo/hello-echo.wasm"))

	grp := rt.NewGroup()
	for i := 0; i < 5000; i++ {
		grp.Add(doWasm([]byte(largeInput)))
	}

	if err := grp.Wait(); err != nil {
		t.Error("group returned an error")
	}
}

func TestWasmLargeDataGroupWithPool(t *testing.T) {
	r := rt.New()

	doWasm := r.Handle("wasm", NewRunner("./testdata/hello-echo/hello-echo.wasm"), rt.PoolSize(5))

	grp := rt.NewGroup()
	for i := 0; i < 5000; i++ {
		grp.Add(doWasm([]byte(largeInput)))
	}

	if err := grp.Wait(); err != nil {
		t.Error("group returned an error")
	}
}

func TestWasmCacheGetSetRustToSwift(t *testing.T) {
	setJob := rt.NewJob("rust-set", "very important")
	getJob := rt.NewJob("swift-get", "")

	_, err := sharedRT.Do(setJob).Then()
	if err != nil {
		t.Error(errors.Wrap(err, "failed to set cache value"))
		return
	}

	r2, err := sharedRT.Do(getJob).Then()
	if err != nil {
		t.Error(errors.Wrap(err, "failed to get cache value"))
		return
	}

	if string(r2.([]byte)) != "very important" {
		t.Error(fmt.Errorf("did not get expected output"))
	}
}

func TestWasmCacheGetSetSwiftToRust(t *testing.T) {
	setJob := rt.NewJob("swift-set", "very important")
	getJob := rt.NewJob("rust-get", "")

	_, err := sharedRT.Do(setJob).Then()
	if err != nil {
		t.Error(errors.Wrap(err, "failed to set cache value"))
		return
	}

	r2, err := sharedRT.Do(getJob).Then()
	if err != nil {
		t.Error(errors.Wrap(err, "failed to get cache value"))
		return
	}

	if string(r2.([]byte)) != "very important" {
		t.Error(fmt.Errorf("did not get expected output"))
	}
}

func TestWasmFileGetStatic(t *testing.T) {
	getJob := rt.NewJob("get-static", "important.md")

	r, err := sharedRT.Do(getJob).Then()
	if err != nil {
		t.Error(errors.Wrap(err, "failed to Do get-static job"))
		return
	}

	result := string(r.([]byte))

	expected := "# Hello, World\n\nContents are very important"

	if result != expected {
		t.Error("failed, got:\n", result, "\nexpeted:\n", expected)
	}
}

func TestWasmFileGetStaticSwift(t *testing.T) {
	getJob := rt.NewJob("get-static-swift", "")

	r, err := sharedRT.Do(getJob).Then()
	if err != nil {
		t.Error(errors.Wrap(err, "failed to Do get-static job"))
		return
	}

	result := string(r.([]byte))

	expected := "# Hello, World\n\nContents are very important"

	if result != expected {
		t.Error("failed, got:\n", result, "\nexpeted:\n", expected)
	}
}
