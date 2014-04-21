package local

import(
    "os"
    "io/ioutil"
    "errors"
    "strings"
)

type Filename struct {
    SubDir string
    Name string
}

func CreateDir(dir string) error {

    if dir == "" {
        return errors.New("create-dir-fail")
    }

    var err = os.Mkdir(dir, 777)
    if err == nil {
        return err
    }

    var n = strings.LastIndex(dir, "/")
    if n == -1 {
        return err
    }

    var dirStr = string([]byte(dir)[0:n])
    err = CreateDir(dirStr)

    if err != nil {
        return err
    }

    err = os.Mkdir(dir, 777)

    return err

}

func FindFile(dir string) ([] string, error) {
    return findFileSub(dir, "")
}

func findFileSub(dir, subDir string) (filenames []string, returnErr error) {

    fileInfoArr, err := ioutil.ReadDir(dir + "/" + subDir)

    if err != nil {
        returnErr = err
        return
    }

    for _, fileInfo := range fileInfoArr {
    
        if fileInfo.IsDir() {
            reqFilenames, err := findFileSub(dir, subDir + fileInfo.Name() + "/")

            if err != nil {
                returnErr = err
                return
            }
            
            filenames = append(filenames, subDir + fileInfo.Name() + "/")
            filenames = append(filenames, reqFilenames...)
        } else {
            filenames = append(filenames, subDir + fileInfo.Name())
        }

    }
    returnErr = nil
    return
}